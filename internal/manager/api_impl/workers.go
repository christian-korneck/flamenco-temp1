package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/pkg/api"
)

// The default BCrypt cost is made for important passwords. For Flamenco, the
// Worker password is not that important.
const bcryptCost = bcrypt.MinCost

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

	logger.Info().Str("nickname", req.Nickname).Msg("registering new worker")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Secret), bcryptCost)
	if err != nil {
		logger.Warn().Err(err).Msg("error hashing worker password")
		return sendAPIError(e, http.StatusBadRequest, "error hashing password")
	}

	dbWorker := persistence.Worker{
		UUID:               uuid.New().String(),
		Name:               req.Nickname,
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
		Nickname:           dbWorker.Name,
		Address:            dbWorker.Address,
		LastActivity:       dbWorker.LastActivity,
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
	w, err := f.workerUpdateAfterSignOn(e, req)
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error storing worker in database")
	}

	resp := api.WorkerStateChange{}
	if w.StatusRequested != "" {
		resp.StatusRequested = w.StatusRequested
	} else {
		resp.StatusRequested = api.WorkerStatusAwake
	}
	return e.JSON(http.StatusOK, resp)
}

func (f *Flamenco) workerUpdateAfterSignOn(e echo.Context, update api.SignOnJSONBody) (*persistence.Worker, error) {
	logger := requestLogger(e)
	w := requestWorkerOrPanic(e)

	// Update the worker for with the new sign-on info.
	w.Status = api.WorkerStatusStarting
	w.Address = e.RealIP()
	w.Name = update.Nickname

	// Remove trailing spaces from task types, and convert to lower case.
	for idx := range update.SupportedTaskTypes {
		update.SupportedTaskTypes[idx] = strings.TrimSpace(strings.ToLower(update.SupportedTaskTypes[idx]))
	}
	w.SupportedTaskTypes = strings.Join(update.SupportedTaskTypes, ",")

	// Save the new Worker info to the database.
	err := f.persist.SaveWorker(e.Request().Context(), w)
	if err != nil {
		logger.Warn().Err(err).
			Str("newStatus", string(w.Status)).
			Msg("error storing Worker in database")
		return nil, err
	}

	return w, nil
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
	w.Status = api.WorkerStatusOffline
	if w.StatusRequested == api.WorkerStatusShutdown {
		w.StatusRequested = ""
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

	// Re-queue all tasks (should be only one) this worker is now working on.
	err = f.workerRequeueActiveTasks(ctx, logger, w)
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error re-queueing your tasks")
	}

	return e.NoContent(http.StatusNoContent)
}

// workerRequeueActiveTasks re-queues all active tasks (should be max one) of this worker.
func (f *Flamenco) workerRequeueActiveTasks(ctx context.Context, logger zerolog.Logger, worker *persistence.Worker) error {
	// Fetch the tasks to update.
	tasks, err := f.persist.FetchTasksOfWorkerInStatus(ctx, worker, api.TaskStatusActive)
	if err != nil {
		return fmt.Errorf("fetching tasks of worker %s in status %q: %w", worker.UUID, api.TaskStatusActive, err)
	}

	// Run each task change through the task state machine.
	var lastErr error
	for _, task := range tasks {
		logger.Info().
			Str("task", task.UUID).
			Msg("re-queueing task")
		err := f.stateMachine.TaskStatusChange(ctx, task, api.TaskStatusQueued)
		if err != nil {
			logger.Warn().Err(err).
				Str("task", task.UUID).
				Msg("error queueing task on worker sign-off")
			lastErr = err
		}
		// TODO: write to task activity that it got requeued because of worker sign-off.
	}

	return lastErr
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

	w.Status = req.Status
	if w.StatusRequested != "" && req.Status != w.StatusRequested {
		logger.Warn().
			Str("workersRequestedStatus", string(w.StatusRequested)).
			Msg("worker changed to status that was not requested")
	} else {
		logger.Info().Msg("worker changed status")
		// Either there was no status change request (and this is a no-op) or the
		// status change was actually acknowledging the request.
		w.StatusRequested = ""
	}

	err = f.persist.SaveWorkerStatus(e.Request().Context(), w)
	if err != nil {
		logger.Warn().Err(err).
			Str("newStatus", string(w.Status)).
			Msg("error storing Worker in database")
		return sendAPIError(e, http.StatusInternalServerError, "error storing worker in database")
	}

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) ScheduleTask(e echo.Context) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)
	logger.Debug().Msg("worker requesting task")

	f.taskSchedulerMutex.Lock()
	defer f.taskSchedulerMutex.Unlock()

	// Check that this worker is actually allowed to do work.
	requiredStatusToGetTask := api.WorkerStatusAwake
	if worker.Status != api.WorkerStatusAwake {
		logger.Warn().
			Str("workerStatus", string(worker.Status)).
			Str("requiredStatus", string(requiredStatusToGetTask)).
			Msg("worker asking for task but is in wrong state")
		return sendAPIError(e, http.StatusConflict,
			fmt.Sprintf("worker is in state %q, requires state %q to execute tasks", worker.Status, requiredStatusToGetTask))
	}
	if worker.StatusRequested != "" {
		logger.Warn().
			Str("workerStatus", string(worker.Status)).
			Str("requestedStatus", string(worker.StatusRequested)).
			Msg("worker asking for task but needs state change first")
		return e.JSON(http.StatusLocked, api.WorkerStateChange{
			StatusRequested: worker.StatusRequested,
		})
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

func (f *Flamenco) TaskUpdate(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)

	if _, err := uuid.Parse(taskID); err != nil {
		logger.Debug().Msg("invalid task ID received")
		return sendAPIError(e, http.StatusBadRequest, "task ID not valid")
	}
	logger = logger.With().Str("taskID", taskID).Logger()

	// Fetch the task, to see if this worker is even allowed to send us updates.
	ctx := e.Request().Context()
	dbTask, err := f.persist.FetchTask(ctx, taskID)
	if err != nil {
		logger.Warn().Err(err).Msg("cannot fetch task")
		if errors.Is(err, persistence.ErrTaskNotFound) {
			return sendAPIError(e, http.StatusNotFound, "task %+v not found", taskID)
		}
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	}
	if dbTask == nil {
		panic("task could not be fetched, but database gave no error either")
	}

	// Decode the request body.
	var taskUpdate api.TaskUpdateJSONRequestBody
	if err := e.Bind(&taskUpdate); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}
	if dbTask.WorkerID == nil {
		logger.Warn().
			Msg("worker trying to update task that's not assigned to any worker")
		return sendAPIError(e, http.StatusConflict, "task %+v is not assigned to any worker, so also not to you", taskID)
	}
	if *dbTask.WorkerID != worker.ID {
		logger.Warn().Msg("worker trying to update task that's assigned to another worker")
		return sendAPIError(e, http.StatusConflict, "task %+v is not assigned to you", taskID)
	}

	// TODO: check whether this task may undergo the requested status change.

	if err := f.doTaskUpdate(ctx, logger, worker, dbTask, taskUpdate); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "unable to handle status update: %v", err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) doTaskUpdate(
	ctx context.Context,
	logger zerolog.Logger,
	w *persistence.Worker,
	dbTask *persistence.Task,
	update api.TaskUpdateJSONRequestBody,
) error {
	if dbTask.Job == nil {
		logger.Panic().Msg("dbTask.Job is nil, unable to continue")
	}

	var dbErr error

	if update.TaskStatus != nil {
		oldTaskStatus := dbTask.Status
		err := f.stateMachine.TaskStatusChange(ctx, dbTask, *update.TaskStatus)
		if err != nil {
			logger.Error().Err(err).
				Str("newTaskStatus", string(*update.TaskStatus)).
				Str("oldTaskStatus", string(oldTaskStatus)).
				Msg("error changing task status")
			dbErr = fmt.Errorf("changing status of task %s to %q: %w",
				dbTask.UUID, *update.TaskStatus, err)
		}
	}

	if update.Activity != nil {
		dbTask.Activity = *update.Activity
		dbErr = f.persist.SaveTaskActivity(ctx, dbTask)
	}

	if update.Log != nil {
		// Errors writing the log to file should be logged in our own logging
		// system, but shouldn't abort the render. As such, `err` is not returned to
		// the caller.
		err := f.logStorage.Write(logger, dbTask.Job.UUID, dbTask.UUID, *update.Log)
		if err != nil {
			logger.Error().Err(err).Msg("error writing task log")
		}
	}

	return dbErr
}

func (f *Flamenco) MayWorkerRun(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)

	if _, err := uuid.Parse(taskID); err != nil {
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

	if mkr.MayKeepRunning {
		// TODO: record that this worker "touched" this task, for timeout calculations.
	}

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
