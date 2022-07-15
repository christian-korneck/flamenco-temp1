package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/internal/manager/last_rendered"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

// RegisterWorker registers a new worker and stores it in the database.
func (f *Flamenco) RegisterWorker(e echo.Context) error {
	logger := requestLogger(e)

	var req api.RegisterWorkerJSONBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	// TODO: validate the request, should at least have non-empty name, secret, and platform.

	logger.Info().Str("name", req.Name).Msg("registering new worker")

	hashedPassword, err := passwordHasher.GenerateHashedPassword([]byte(req.Secret))
	if err != nil {
		logger.Warn().Err(err).Msg("error hashing worker password")
		return sendAPIError(e, http.StatusBadRequest, "error hashing password")
	}

	dbWorker := persistence.Worker{
		UUID:               uuid.New(),
		Name:               req.Name,
		Secret:             string(hashedPassword),
		Platform:           req.Platform,
		Address:            e.RealIP(),
		SupportedTaskTypes: strings.Join(req.SupportedTaskTypes, ","),
	}
	if err := f.persist.CreateWorker(e.Request().Context(), &dbWorker); err != nil {
		logger.Warn().Err(err).Msg("error creating new worker in DB")
		if persistence.ErrIsDBBusy(err) {
			return sendAPIErrorDBBusy(e, "too busy to register worker, try again later")
		}
		return sendAPIError(e, http.StatusBadRequest, "error registering worker")
	}

	return e.JSON(http.StatusOK, &api.RegisteredWorker{
		Uuid:               dbWorker.UUID,
		Name:               dbWorker.Name,
		Address:            dbWorker.Address,
		Platform:           dbWorker.Platform,
		Software:           dbWorker.Software,
		Status:             dbWorker.Status,
		SupportedTaskTypes: strings.Split(dbWorker.SupportedTaskTypes, ","),
	})
}

func (f *Flamenco) SignOn(e echo.Context) error {
	logger := requestLogger(e)

	var req api.SignOnJSONBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	logger.Info().Msg("worker signing on")
	w, prevStatus, err := f.workerUpdateAfterSignOn(e, req)
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error storing worker in database")
	}

	// Broadcast the status change.
	update := webupdates.NewWorkerUpdate(w)
	if prevStatus != "" {
		update.PreviousStatus = &prevStatus
	}
	f.broadcaster.BroadcastWorkerUpdate(update)

	resp := api.WorkerStateChange{}
	if w.StatusRequested != "" {
		resp.StatusRequested = w.StatusRequested
	} else {
		resp.StatusRequested = api.WorkerStatusAwake
	}

	return e.JSON(http.StatusOK, resp)
}

func (f *Flamenco) workerUpdateAfterSignOn(e echo.Context, update api.SignOnJSONBody) (*persistence.Worker, api.WorkerStatus, error) {
	logger := requestLogger(e)
	w := requestWorkerOrPanic(e)
	ctx := e.Request().Context()

	// Update the worker for with the new sign-on info.
	prevStatus := w.Status
	w.Status = api.WorkerStatusStarting
	w.Address = e.RealIP()
	w.Name = update.Name
	w.Software = update.SoftwareVersion

	// Remove trailing spaces from task types, and convert to lower case.
	for idx := range update.SupportedTaskTypes {
		update.SupportedTaskTypes[idx] = strings.TrimSpace(strings.ToLower(update.SupportedTaskTypes[idx]))
	}
	w.SupportedTaskTypes = strings.Join(update.SupportedTaskTypes, ",")

	// Save the new Worker info to the database.
	err := f.persist.SaveWorker(ctx, w)
	if err != nil {
		logger.Warn().Err(err).
			Str("newStatus", string(w.Status)).
			Msg("error storing Worker in database")
		return nil, "", err
	}

	err = f.workerSeen(ctx, logger, w)
	if err != nil {
		return nil, "", err
	}

	return w, prevStatus, nil
}

func (f *Flamenco) SignOff(e echo.Context) error {
	logger := requestLogger(e)

	var req api.SignOnJSONBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	logger.Info().Msg("worker signing off")
	w := requestWorkerOrPanic(e)
	prevStatus := w.Status
	w.Status = api.WorkerStatusOffline
	if w.StatusRequested == api.WorkerStatusOffline {
		w.StatusChangeClear()
	}

	// Pass a generic background context, as these changes should be stored even
	// when the HTTP connection is aborted.
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	err = f.persist.SaveWorkerStatus(ctx, w)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("newStatus", string(w.Status)).
			Msg("error storing worker status in database")
		return sendAPIError(e, http.StatusInternalServerError, "error storing new status in database")
	}

	// Ignore database errors here; the rest of the signoff process should just happen.
	_ = f.workerSeen(ctx, logger, w)

	// Re-queue all tasks (should be only one) this worker is now working on.
	err = f.stateMachine.RequeueActiveTasksOfWorker(ctx, w, "worker signed off")
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error re-queueing your tasks")
	}

	update := webupdates.NewWorkerUpdate(w)
	update.PreviousStatus = &prevStatus
	f.broadcaster.BroadcastWorkerUpdate(update)

	return e.NoContent(http.StatusNoContent)
}

// (GET /api/worker/state)
func (f *Flamenco) WorkerState(e echo.Context) error {
	worker := requestWorkerOrPanic(e)

	if worker.StatusRequested == "" {
		return e.NoContent(http.StatusNoContent)
	}

	return e.JSON(http.StatusOK, api.WorkerStateChange{
		StatusRequested: worker.StatusRequested,
	})
}

// Worker changed state. This could be as acknowledgement of a Manager-requested state change, or in response to worker-local signals.
// (POST /api/worker/state-changed)
func (f *Flamenco) WorkerStateChanged(e echo.Context) error {
	logger := requestLogger(e)

	var req api.WorkerStateChangedJSONRequestBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	w := requestWorkerOrPanic(e)
	logger = logger.With().
		Str("currentStatus", string(w.Status)).
		Str("newStatus", string(req.Status)).
		Logger()

	prevStatus := w.Status
	w.Status = req.Status
	if w.StatusRequested != "" && req.Status != w.StatusRequested {
		logger.Warn().
			Str("workersRequestedStatus", string(w.StatusRequested)).
			Msg("worker changed to status that was not requested")
	} else {
		logger.Info().Msg("worker changed status")
		// Either there was no status change request (and this is a no-op) or the
		// status change was actually acknowledging the request.
		w.StatusChangeClear()
	}

	ctx := e.Request().Context()
	err = f.persist.SaveWorkerStatus(ctx, w)
	if err != nil {
		logger.Warn().Err(err).
			Str("newStatus", string(w.Status)).
			Msg("error storing Worker in database")
		return sendAPIError(e, http.StatusInternalServerError, "error storing worker in database")
	}

	if err := f.workerSeen(ctx, logger, w); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error storing worker 'last seen' timestamp in database")
	}

	update := webupdates.NewWorkerUpdate(w)
	update.PreviousStatus = &prevStatus
	f.broadcaster.BroadcastWorkerUpdate(update)

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) ScheduleTask(e echo.Context) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)
	ctx := e.Request().Context()
	logger.Debug().Msg("worker requesting task")

	f.taskSchedulerMutex.Lock()
	defer f.taskSchedulerMutex.Unlock()

	// The worker is actively asking for a task, so note that it was seen
	// regardless of any failures below, or whether there actually is a task to
	// run.
	if err := f.workerSeen(ctx, logger, worker); err != nil {
		return sendAPIError(e, http.StatusInternalServerError,
			"error storing worker 'last seen' timestamp in database")
	}

	// Check that this worker is actually allowed to do work.
	if worker.StatusRequested != "" {
		logger.Warn().
			Str("workerStatus", string(worker.Status)).
			Str("requestedStatus", string(worker.StatusRequested)).
			Msg("worker asking for task but needs state change first")
		return e.JSON(http.StatusLocked, api.WorkerStateChange{
			StatusRequested: worker.StatusRequested,
		})
	}

	requiredStatusToGetTask := api.WorkerStatusAwake
	if worker.Status != requiredStatusToGetTask {
		logger.Warn().
			Str("workerStatus", string(worker.Status)).
			Str("requiredStatus", string(requiredStatusToGetTask)).
			Msg("worker asking for task but is in wrong state")
		return sendAPIError(e, http.StatusConflict,
			fmt.Sprintf("worker is in state %q, requires state %q to execute tasks", worker.Status, requiredStatusToGetTask))
	}

	// Get a task to execute:
	dbTask, err := f.persist.ScheduleTask(e.Request().Context(), worker)
	if err != nil {
		if persistence.ErrIsDBBusy(err) {
			logger.Warn().Msg("database busy scheduling task for worker")
			return sendAPIErrorDBBusy(e, "too busy to find a task for you, try again later")
		}
		logger.Warn().Err(err).Msg("error scheduling task for worker")
		return sendAPIError(e, http.StatusInternalServerError, "internal error finding a task for you: %v", err)
	}
	if dbTask == nil {
		return e.NoContent(http.StatusNoContent)
	}

	// Add a note to the task log about the worker assignment.
	msg := fmt.Sprintf("Task assigned to worker %s (%s)", worker.Name, worker.UUID)
	if err := f.logStorage.WriteTimestamped(logger, dbTask.Job.UUID, dbTask.UUID, msg); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "internal error appending to task log: %v", err)
	}

	// Move the task to 'active' status so that it won't be assigned to another
	// worker. This also enables the task timeout monitoring.
	if err := f.stateMachine.TaskStatusChange(ctx, dbTask, api.TaskStatusActive); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "internal error marking task as active: %v", err)
	}

	// Start timeout measurement as soon as the Worker gets the task assigned.
	if err := f.workerPingedTask(ctx, logger, dbTask); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "internal error updating task for timeout calculation: %v", err)
	}

	// Convert database objects to API objects:
	apiCommands := []api.Command{}
	for _, cmd := range dbTask.Commands {
		apiCommands = append(apiCommands, api.Command{
			Name:       cmd.Name,
			Parameters: cmd.Parameters,
		})
	}
	apiTask := api.AssignedTask{
		Uuid:        dbTask.UUID,
		Commands:    apiCommands,
		Job:         dbTask.Job.UUID,
		JobPriority: dbTask.Job.Priority,
		JobType:     dbTask.Job.JobType,
		Name:        dbTask.Name,
		Priority:    dbTask.Priority,
		Status:      api.TaskStatus(dbTask.Status),
		TaskType:    dbTask.Type,
	}

	// Perform variable replacement before sending to the Worker.
	customisedTask := replaceTaskVariables(f.config, apiTask, *worker)
	return e.JSON(http.StatusOK, customisedTask)
}

func (f *Flamenco) TaskOutputProduced(e echo.Context, taskID string) error {
	ctx := e.Request().Context()
	filesize := e.Request().ContentLength
	worker := requestWorkerOrPanic(e)
	logger := requestLogger(e).With().
		Str("task", taskID).
		Int64("imageSizeBytes", filesize).
		Logger()

	err := f.workerSeen(ctx, logger, worker)
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error updating 'last seen' timestamp of worker: %v", err)
	}

	// Check the file size:
	switch {
	case filesize <= 0:
		logger.Warn().Msg("TaskOutputProduced: Worker did not sent Content-Length header")
		return sendAPIError(e, http.StatusLengthRequired, "Content-Length header required")
	case filesize > last_rendered.MaxImageSizeBytes:
		logger.Warn().
			Int64("imageSizeBytesMax", last_rendered.MaxImageSizeBytes).
			Msg("TaskOutputProduced: Worker sent too large last-rendered image")
		return sendAPIError(e, http.StatusRequestEntityTooLarge,
			"image too large; should be max %v bytes", last_rendered.MaxImageSizeBytes)
	}

	// Fetch the task, to find its job UUID:
	dbTask, err := f.persist.FetchTask(ctx, taskID)
	switch {
	case errors.Is(err, persistence.ErrTaskNotFound):
		return e.JSON(http.StatusNotFound, "Task does not exist")
	case err != nil:
		logger.Error().Err(err).Msg("TaskOutputProduced: cannot fetch task")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	case dbTask == nil:
		panic("task could not be fetched, but database gave no error either")
	}

	// Read the image bytes into memory.
	imageBytes, err := io.ReadAll(e.Request().Body)
	if err != nil {
		logger.Warn().Err(err).Msg("TaskOutputProduced: error reading image from request")
		return sendAPIError(e, http.StatusBadRequest, "error reading request body: %v", err)
	}

	// Create the "last rendered" payload.
	jobUUID := dbTask.Job.UUID
	thumbnailInfo, err := f.lastRenderedInfoForJob(logger, jobUUID)
	if err != nil {
		logger.Error().Err(err).Msg("TaskOutputProduced: error getting last-rendered thumbnail info for job")
		return sendAPIError(e, http.StatusInternalServerError, "error getting last-rendered thumbnail info for job: %v", err)
	}
	payload := last_rendered.Payload{
		JobUUID:    jobUUID,
		WorkerUUID: worker.UUID,
		MimeType:   e.Request().Header.Get("Content-Type"),
		Image:      imageBytes,

		Callback: func(ctx context.Context) {
			// Store this job as the last one to get a rendered image.
			f.persist.SetLastRendered(ctx, dbTask.Job)

			// Broadcast when the processing is done.
			update := webupdates.NewLastRenderedUpdate(jobUUID)
			update.Thumbnail = *thumbnailInfo
			f.broadcaster.BroadcastLastRenderedImage(update)
		},
	}

	// Queue the image for processing:
	err = f.lastRender.QueueImage(payload)
	if err != nil {
		switch {
		case errors.Is(err, last_rendered.ErrMimeTypeUnsupported):
			logger.Warn().
				Str("mimeType", payload.MimeType).
				Msg("TaskOutputProduced: Worker sent unsupported mime type")
			return sendAPIError(e, http.StatusUnsupportedMediaType, "unsupported mime type %q", payload.MimeType)
		case errors.Is(err, last_rendered.ErrQueueFull):
			logger.Info().
				Msg("TaskOutputProduced: image processing queue is full, ignoring request")
			return sendAPIError(e, http.StatusTooManyRequests, "image processing queue is full")
		default:
			logger.Error().Err(err).
				Msg("TaskOutputProduced: error queueing image")
			return sendAPIError(e, http.StatusInternalServerError, "error queueing image for processing: %v", err)
		}
	}

	logger.Info().Msg("TaskOutputProduced: accepted last-rendered image for processing")
	return e.NoContent(http.StatusAccepted)
}

func (f *Flamenco) workerPingedTask(
	ctx context.Context,
	logger zerolog.Logger,
	task *persistence.Task,
) error {
	err := f.persist.TaskTouchedByWorker(ctx, task)
	if err != nil {
		logger.Error().Err(err).Msg("error marking task as 'touched' by worker")
		return err
	}
	return nil
}

// workerSeen marks the worker as 'seen' and logs any database error that may occur.
func (f *Flamenco) workerSeen(
	ctx context.Context,
	logger zerolog.Logger,
	w *persistence.Worker,
) error {
	err := f.persist.WorkerSeen(ctx, w)
	if err != nil {
		logger.Error().Err(err).Msg("error marking Worker as 'seen' in the database")
		return err
	}
	return nil
}

func (f *Flamenco) MayWorkerRun(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)

	if !uuid.IsValid(taskID) {
		logger.Debug().Msg("invalid task ID received")
		return sendAPIError(e, http.StatusBadRequest, "task ID not valid")
	}
	logger = logger.With().Str("task", taskID).Logger()

	// Lock the task scheduler so that tasks don't get reassigned while we perform our checks.
	f.taskSchedulerMutex.Lock()
	defer f.taskSchedulerMutex.Unlock()

	// Fetch the task, to see if this worker is allowed to run it.
	ctx := e.Request().Context()
	dbTask, err := f.persist.FetchTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, persistence.ErrTaskNotFound) {
			mkr := api.MayKeepRunning{Reason: "Task not found"}
			return e.JSON(http.StatusOK, mkr)
		}
		logger.Error().Err(err).Msg("MayWorkerRun: cannot fetch task")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	}
	if dbTask == nil {
		panic("task could not be fetched, but database gave no error either")
	}

	mkr := mayWorkerRun(worker, dbTask)

	// Errors saving the "worker pinged task" and "worker seen" fields in the
	// database are just logged. It's not something to bother the worker with.
	if mkr.MayKeepRunning {
		_ = f.workerPingedTask(ctx, logger, dbTask)
	}
	_ = f.workerSeen(ctx, logger, worker)

	return e.JSON(http.StatusOK, mkr)
}

// mayWorkerRun checks the worker and the task, to see if this worker may keep running this task.
func mayWorkerRun(worker *persistence.Worker, dbTask *persistence.Task) api.MayKeepRunning {
	if worker.StatusRequested != "" {
		return api.MayKeepRunning{
			Reason:                "worker status change requested",
			StatusChangeRequested: true,
		}
	}
	if dbTask.WorkerID == nil || *dbTask.WorkerID != worker.ID {
		return api.MayKeepRunning{Reason: "task not assigned to this worker"}
	}
	if !task_state_machine.IsRunnableTaskStatus(dbTask.Status) {
		return api.MayKeepRunning{Reason: fmt.Sprintf("task is in non-runnable status %q", dbTask.Status)}
	}
	return api.MayKeepRunning{MayKeepRunning: true}
}
