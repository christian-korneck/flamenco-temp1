import logging

from bson import ObjectId
from flask import Blueprint, request, url_for
import werkzeug.exceptions as wz_exceptions
import werkzeug.datastructures

from pillar import current_app
from pillar.api.utils import authorization, authentication, utcnow, random_etag, str2id, jsonify

from flamenco import current_flamenco

api_blueprint = Blueprint('flamenco.managers.api', __name__)
log = logging.getLogger(__name__)

# Task statuses that are acceptable after a task has been set to 'cancel-requested'
# TODO: maybe move allowed task transition handling to a different bit of code.
ACCEPTED_AFTER_CANCEL_REQUESTED = {'canceled', 'failed', 'completed'}

DEPSGRAPH_RUNNABLE_JOB_STATUSES = ['queued', 'active', 'cancel-requested', 'soft-failed']
DEPSGRAPH_CLEAN_SLATE_TASK_STATUSES = ['queued', 'claimed-by-manager',
                                       'active', 'cancel-requested', 'soft-failed']
DEPSGRAPH_MODIFIED_SINCE_TASK_STATUSES = ['queued', 'claimed-by-manager', 'soft-failed']

# Number of lines of logging to keep on the task itself.
LOG_TAIL_LINES = 10


def manager_api_call(*args, pass_manager_doc=False):
    """Decorator, performs some standard stuff for Manager API endpoints.

    :param pass_manager_doc: causes either the manager ID (False) or entire
        document (True) to be passed as the first parameter to the decorated
        function.
    """

    if args and hasattr(args[0], '__call__'):
        raise TypeError('use @manager_api_call() <-- note the parentheses')

    def decorator(wrapped):
        import functools

        @authorization.require_login(require_roles={'service', 'flamenco_manager'},
                                     require_all=True)
        @functools.wraps(wrapped)
        def wrapper(manager_id, *args, **kwargs):
            from pillar.api.utils import mongo

            manager_id = str2id(manager_id)
            manager = mongo.find_one_or_404('flamenco_managers', manager_id)
            if not current_flamenco.manager_manager.user_manages(mngr_doc=manager):
                user_id = authentication.current_user_id()
                log.warning(
                    'Service account %s called manager API endpoint for manager %s of another '
                    'service account', user_id, manager_id)
                raise wz_exceptions.Unauthorized()

            return wrapped(manager if pass_manager_doc else manager_id,
                           request.json,
                           *args, **kwargs)

        return wrapper

    return decorator


@api_blueprint.route('/<manager_id>/startup', methods=['POST'])
@manager_api_call()
def startup(manager_id, notification):
    log.info('Received startup notification from manager %s %s', manager_id, notification)
    return handle_notification(manager_id, notification)


@api_blueprint.route('/<manager_id>/update', methods=['POST'])
@manager_api_call()
def update(manager_id, notification):
    log.info('Received configuration update notification from manager %s %s',
             manager_id, notification)
    return handle_notification(manager_id, notification)


def handle_notification(manager_id: str, notification: dict):
    """Handle startup and update notifications.

    These notifications contain info about the Manager, including the task
    types supported by its Workers.
    """
    from flamenco import current_flamenco
    import uuid
    import datetime

    if not notification:
        raise wz_exceptions.BadRequest('no JSON payload received')

    settings_version = notification.get('_meta', {}).get('version', 1)
    updates_unset = {}
    try:
        updates_set = {
            '_updated': datetime.datetime.utcnow(),
            '_etag': uuid.uuid4().hex,
            'url': notification['manager_url'],
            'settings_version': settings_version,
            'variables': notification['variables'],
            'stats.nr_of_workers': notification['nr_of_workers'],
        }
        if settings_version <= 1:
            updates_set['path_replacement'] = notification['path_replacement']
        else:
            updates_unset['path_replacement'] = True
    except KeyError as ex:
        raise wz_exceptions.BadRequest(f'Missing key {ex}')

    try:
        updates_set['worker_task_types'] = notification['worker_task_types']
    except KeyError:
        pass

    mngr_coll = current_flamenco.db('managers')
    updates = {'$set': updates_set}
    if updates_unset:
        updates['$unset'] = updates_unset

    update_res = mngr_coll.update_one({'_id': manager_id}, updates)
    if update_res.matched_count != 1:
        log.warning('Updating manager %s matched %i documents.',
                    manager_id, update_res.matched_count)
        raise wz_exceptions.InternalServerError('Unable to update manager in database.')

    return '', 204


@api_blueprint.route('/<manager_id>/task-update-batch', methods=['POST'])
@manager_api_call(pass_manager_doc=True)
def task_update_batch(manager: dict, task_updates):
    """Handle task updates from the Manager and respond with actions for the Manager.

    This endpoint receives batched task updates from the Manager, and handles
    those (for example by failing a job when too many tasks failed).

    In the response there are a few other pieces of data that basically
    indicate actions to be performed by the manager (cancelling tasks, sending
    task log files).
    """
    from pillar.api.utils import jsonify

    manager_id = manager['_id']
    total_modif_count, handled_update_ids = handle_task_update_batch(manager_id, task_updates)

    # Check which tasks are in state 'cancel-requested', as those need to be sent back.
    # This MUST be done after we run the task update batch, as just-changed task statuses
    # should be taken into account.
    tasks_to_cancel = tasks_cancel_requested(manager_id)

    response = {'modified_count': total_modif_count,
                'handled_update_ids': handled_update_ids}
    if tasks_to_cancel:
        response['cancel_task_ids'] = list(tasks_to_cancel)

    # TODO(Sybren): expose the fact that there are other types of actions to
    # perform by the Manager in a different way, for example via a Redis or
    # RabbitMQ channel.
    upload_task_file_queue = manager.get('upload_task_file_queue')
    if upload_task_file_queue:
        response['upload_task_file_queue'] = upload_task_file_queue

    return jsonify(response)


def handle_task_update_batch(manager_id, task_updates):
    """Performs task updates.

    Task status changes are generally always accepted. The only exception is when the
    task ID is contained in 'tasks_to_cancel'; in that case only a transition to either
    'canceled', 'completed' or 'failed' is accepted.

    :returns: tuple (total nr of modified tasks, handled update IDs)
    """

    if not task_updates:
        return 0, []

    import dateutil.parser
    from pillar.api.utils import str2id

    from flamenco import current_flamenco, eve_settings

    log.debug('Received %i task updates from manager %s', len(task_updates), manager_id)

    tasks_coll = current_flamenco.db('tasks')
    logs_coll = current_flamenco.db('task_logs')

    valid_statuses = set(eve_settings.tasks_schema['status']['allowed'])
    handled_update_ids = []
    total_modif_count = 0

    for task_update in task_updates:
        # Check that this task actually belongs to this manager, before we accept any updates.
        update_id = str2id(task_update['_id'])
        task_id = str2id(task_update['task_id'])
        task_info = tasks_coll.find_one({'_id': task_id},
                                        projection={'manager': 1, 'status': 1, 'job': 1})

        # For now, we just ignore updates to non-existing tasks. Someone might have just deleted
        # one, for example. This is not a reason to reject the entire batch.
        if task_info is None:
            log.warning('Manager %s sent update for non-existing task %s; accepting but ignoring',
                        manager_id, task_id)
            handled_update_ids.append(update_id)
            continue

        if task_info['manager'] != manager_id:
            log.info('Manager %s sent update for task %s which belongs to other manager %s',
                        manager_id, task_id, task_info['manager'])
            continue

        if task_update.get('received_on_manager'):
            received_on_manager = dateutil.parser.parse(task_update['received_on_manager'])
        else:
            # Fake a 'received on manager' field; it really should have been in the JSON payload.
            received_on_manager = utcnow()

        # Store the log for this task, allowing for duplicate log reports.
        #
        # NOTE: is deprecated and will be removed in a future version of Flamenco;
        # only periodically send the last few lines of logging in 'log_tail' and
        # store the entire log on the Manager itself.
        task_log = task_update.get('log')
        if task_log:
            log_doc = {
                '_id': update_id,
                'task': task_id,
                'received_on_manager': received_on_manager,
                'log': task_log
            }
            logs_coll.replace_one({'_id': update_id}, log_doc, upsert=True)

        # Modify the task, and append the log to the logs collection.
        updates = {
            'task_progress_percentage': task_update.get('task_progress_percentage', 0),
            'current_command_index': task_update.get('current_command_index', 0),
            'command_progress_percentage': task_update.get('command_progress_percentage', 0),
            '_updated': received_on_manager,
            '_etag': random_etag(),
        }

        new_status = determine_new_task_status(manager_id, task_id, task_info,
                                               task_update.get('task_status'), valid_statuses)
        if new_status:
            updates['status'] = new_status

        new_activity = task_update.get('activity')
        if new_activity:
            updates['activity'] = new_activity
        worker = task_update.get('worker')
        if worker:
            updates['worker'] = worker
        metrics_timing = task_update.get('metrics', {}).get('timing')
        if metrics_timing:
            updates['metrics.timing'] = metrics_timing

        fbw = task_update.get('failed_by_workers')
        if fbw is not None:
            updates['failed_by_workers'] = fbw

        # Store the last lines of logging on the task itself.
        task_log_tail: str = task_update.get('log_tail')
        if not task_log_tail and task_log:
            task_log_tail = '\n'.join(task_log.split('\n')[-LOG_TAIL_LINES:])
        if task_log_tail:
            updates['log'] = task_log_tail

        result = tasks_coll.update_one({'_id': task_id}, {'$set': updates})
        total_modif_count += result.modified_count

        handled_update_ids.append(update_id)

        # Update the task's job after updating the task itself.
        if new_status:
            current_flamenco.job_manager.update_job_after_task_status_change(
                task_info['job'], task_id, new_status)

    return total_modif_count, handled_update_ids


def determine_new_task_status(manager_id, task_id, current_task_info, new_status, valid_statuses):
    """Returns the new task status, or None if the task should not get a new status."""

    if not new_status:
        return None

    current_status = current_task_info['status']
    if new_status == current_status:
        return None

    if current_status == 'cancel-requested':
        if new_status not in ACCEPTED_AFTER_CANCEL_REQUESTED:
            log.info('Manager %s wants to set task %s to status %r, but that is not allowed '
                     'because the task is in status %s',
                     manager_id, task_id, new_status, current_status)
            return None

    if new_status not in valid_statuses:
        # We have to accept the invalid status, because we're too late in the update
        # pipeline to do anything about it. The alternative is to drop the update or
        # reject the entire batch of updates, which is more damaging to the workflow.
        log.warning('Manager %s sent update for task %s with invalid status %r',
                    manager_id, task_id, new_status)
        return None

    return new_status


def tasks_cancel_requested(manager_id):
    """Returns a set of tasks of status cancel-requested."""

    from flamenco import current_flamenco, eve_settings

    tasks_coll = current_flamenco.db('tasks')

    task_ids = {
        task['_id']
        for task in tasks_coll.find({'manager': manager_id, 'status': 'cancel-requested'},
                                    projection={'_id': 1})
    }

    log.debug('Returning %i tasks to be canceled by manager %s', len(task_ids), manager_id)
    return task_ids


@api_blueprint.route('/<manager_id>/depsgraph')
@manager_api_call()
def get_depsgraph(manager_id, request_json):
    """Returns the dependency graph of all tasks assigned to the given Manager.

    Use the HTTP header X-Flamenco-If-Updated-Since to limit the dependency
    graph to tasks that have been modified since that timestamp.
    """

    import dateutil.parser
    from pillar.api.utils import jsonify, bsonify
    from flamenco import current_flamenco
    from flamenco.utils import report_duration

    modified_since = request.headers.get('X-Flamenco-If-Updated-Since')

    with report_duration(log, 'depsgraph query'):
        tasks_coll = current_flamenco.db('tasks')
        jobs_coll = current_flamenco.db('jobs')

        # Get runnable jobs first, as non-runnable jobs are not interesting.
        # Note that jobs going from runnable to non-runnable should have their
        # tasks set to cancel-requested, which is communicated to the Manager
        # through a different channel.
        jobs = jobs_coll.find({
            'manager': manager_id,
            'status': {'$in': DEPSGRAPH_RUNNABLE_JOB_STATUSES}},
            projection={'_id': 1},
        )
        job_ids = [job['_id'] for job in jobs]
        if not job_ids:
            log.debug('Returning empty depsgraph')
            return '', 204  # empty response

        log.debug('Requiring jobs to be in %s', job_ids)
        task_query = {
            'manager': manager_id,
            'status': {'$nin': ['active']},
            'job': {'$in': job_ids},
        }

        if modified_since is None:
            # "Clean slate" query.
            task_query['status'] = {'$in': DEPSGRAPH_CLEAN_SLATE_TASK_STATUSES}
        else:
            # Not clean slate, just give all updated tasks assigned to this manager.
            log.debug('Modified-since header: %s', modified_since)
            modified_since = dateutil.parser.parse(modified_since)
            task_query['_updated'] = {'$gt': modified_since}
            task_query['status'] = {'$in': DEPSGRAPH_MODIFIED_SINCE_TASK_STATUSES}
            log.debug('Querying all tasks changed since %s', modified_since)

        cursor = tasks_coll.find(task_query)
        depsgraph = list(cursor)

    if len(depsgraph) == 0:
        log.debug('Returning empty depsgraph')
        if modified_since is not None:
            return '', 304  # Not Modified
    else:
        log.info('Returning depsgraph of %i tasks', len(depsgraph))

    # Update the task status in the database to move queued tasks to claimed-by-manager.
    # This also erases the link to any previously uploaded log files, to ensure the
    # log is fresh and represents the current execution of the task.
    task_query['status'] = 'queued'
    tasks_coll.update_many(task_query, {
        '$set': {'status': 'claimed-by-manager'},
        '$unset': {'log_file': True},
    })

    # Update the returned task statuses. Unfortunately Mongo doesn't support
    # find_and_modify() on multiple documents.
    for task in depsgraph:
        if task['status'] == 'queued':
            task['status'] = 'claimed-by-manager'

    # Must be a dict to convert to BSON.
    respdoc = {
        'depsgraph': depsgraph,
    }
    if request.accept_mimetypes.best == 'application/bson':
        resp = bsonify(respdoc)
    else:
        resp = jsonify(respdoc)

    if depsgraph:
        last_modification = max(task['_updated'] for task in depsgraph)
        log.debug('Last modification was %s', last_modification)
        # We need a format that can handle sub-second precision, which is not provided by the
        # HTTP date format (RFC 1123). This means that we can't use the Last-Modified header, as
        # it may be incorrectly interpreted and rewritten by HaProxy, Apache or other software
        # in the path between client & server.
        resp.headers['X-Flamenco-Last-Updated'] = last_modification.isoformat()
        resp.headers['X-Flamenco-Last-Updated-Format'] = 'ISO-8601'
    return resp


@api_blueprint.route('/<manager_id>/attach-task-log/<task_id>', methods=['POST'])
@manager_api_call()
def attach_task_log(manager_id: ObjectId, _, task_id: str):
    """Store the POSTed task log as a file in the storage backend.

    Also updates the task itself to have a reference to the file.
    """
    # We only want to deal with GZipped files.
    if 'logfile' not in request.files:
        raise wz_exceptions.BadRequest("Missing uploaded file named 'logfile'")
    uploaded_file: werkzeug.datastructures.FileStorage = request.files['logfile']
    if not uploaded_file.filename.endswith('.gz'):
        # The test HTTP client doesn't support setting per-part headers.
        raise wz_exceptions.BadRequest(f'GZIP your file!')

    # De-queue now; if the task or project doesn't exist, the Manager shouldn't be asked again.
    task_oid = str2id(task_id)
    current_flamenco.manager_manager.dequeue_task_log_request(manager_id, task_oid)

    # Check whether this Manager may attach to this Task.
    tasks_coll = current_flamenco.db('tasks')
    task = tasks_coll.find_one({'_id': task_oid, 'manager': manager_id})
    if not task:
        raise wz_exceptions.NotFound(f'No such task exists')

    proj_coll = current_app.db('projects')
    project = proj_coll.find_one({'_id': task['project'], '_deleted': {'$ne': True}},
                                 projection={'url': True})
    if not project:
        log.warning('attach_task_log(%s, %s): project %s does not exist!',
                    manager_id, task_id, task['project'])
        raise wz_exceptions.NotFound(f'Project for task {task_oid} does not exist')

    preexisting = current_flamenco.task_manager.api_attach_log(task, uploaded_file)

    resp = jsonify({'_message': 'ok'}, status=200 if preexisting else 201)
    resp.headers['Location'] = url_for(
        'flamenco.tasks.perproject.download_task_log_file',
        project_url=project['url'], task_id=task_id)
    return resp


def setup_app(app):
    app.register_api_blueprint(api_blueprint, url_prefix='/flamenco/managers')
