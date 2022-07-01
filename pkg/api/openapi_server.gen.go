// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get the configuration of this Manager.
	// (GET /api/configuration)
	GetConfiguration(ctx echo.Context) error
	// Submit a new job for Flamenco Manager to execute.
	// (POST /api/jobs)
	SubmitJob(ctx echo.Context) error
	// Get the URL that serves the last-rendered images.
	// (GET /api/jobs/last-rendered)
	FetchGlobalLastRenderedInfo(ctx echo.Context) error
	// Fetch list of jobs.
	// (POST /api/jobs/query)
	QueryJobs(ctx echo.Context) error
	// Get single job type and its parameters.
	// (GET /api/jobs/type/{typeName})
	GetJobType(ctx echo.Context, typeName string) error
	// Get list of job types and their parameters.
	// (GET /api/jobs/types)
	GetJobTypes(ctx echo.Context) error
	// Fetch info about the job.
	// (GET /api/jobs/{job_id})
	FetchJob(ctx echo.Context, jobId string) error
	// Remove entries from a job blocklist.
	// (DELETE /api/jobs/{job_id}/blocklist)
	RemoveJobBlocklist(ctx echo.Context, jobId string) error
	// Fetch the list of workers that are blocked from doing certain task types on this job.
	// (GET /api/jobs/{job_id}/blocklist)
	FetchJobBlocklist(ctx echo.Context, jobId string) error
	// Get the URL that serves the last-rendered images of this job.
	// (GET /api/jobs/{job_id}/last-rendered)
	FetchJobLastRenderedInfo(ctx echo.Context, jobId string) error

	// (POST /api/jobs/{job_id}/setstatus)
	SetJobStatus(ctx echo.Context, jobId string) error
	// Fetch a summary of all tasks of the given job.
	// (GET /api/jobs/{job_id}/tasks)
	FetchJobTasks(ctx echo.Context, jobId string) error
	// Fetch a single task.
	// (GET /api/tasks/{task_id})
	FetchTask(ctx echo.Context, taskId string) error
	// Fetch the last few lines of the task's log.
	// (GET /api/tasks/{task_id}/logtail)
	FetchTaskLogTail(ctx echo.Context, taskId string) error

	// (POST /api/tasks/{task_id}/setstatus)
	SetTaskStatus(ctx echo.Context, taskId string) error
	// Get the Flamenco version of this Manager
	// (GET /api/version)
	GetVersion(ctx echo.Context) error
	// Get list of workers.
	// (GET /api/worker-mgt/workers)
	FetchWorkers(ctx echo.Context) error
	// Fetch info about the worker.
	// (GET /api/worker-mgt/workers/{worker_id})
	FetchWorker(ctx echo.Context, workerId string) error

	// (POST /api/worker-mgt/workers/{worker_id}/setstatus)
	RequestWorkerStatusChange(ctx echo.Context, workerId string) error
	// Register a new worker
	// (POST /api/worker/register-worker)
	RegisterWorker(ctx echo.Context) error
	// Mark the worker as offline
	// (POST /api/worker/sign-off)
	SignOff(ctx echo.Context) error
	// Authenticate & sign in the worker.
	// (POST /api/worker/sign-on)
	SignOn(ctx echo.Context) error

	// (GET /api/worker/state)
	WorkerState(ctx echo.Context) error
	// Worker changed state. This could be as acknowledgement of a Manager-requested state change, or in response to worker-local signals.
	// (POST /api/worker/state-changed)
	WorkerStateChanged(ctx echo.Context) error
	// Obtain a new task to execute
	// (POST /api/worker/task)
	ScheduleTask(ctx echo.Context) error
	// Update the task, typically to indicate progress, completion, or failure.
	// (POST /api/worker/task/{task_id})
	TaskUpdate(ctx echo.Context, taskId string) error
	// The response indicates whether the worker is allowed to run / keep running the task. Optionally contains a queued worker status change.
	// (GET /api/worker/task/{task_id}/may-i-run)
	MayWorkerRun(ctx echo.Context, taskId string) error
	// Store the most recently rendered frame here. Note that it is up to the Worker to ensure this is in a format that's digestable by the Manager. Currently only PNG and JPEG support is planned.
	// (POST /api/worker/task/{task_id}/output-produced)
	TaskOutputProduced(ctx echo.Context, taskId string) error
	// Create a directory, and symlink the required files into it. The files must all have been uploaded to Shaman before calling this endpoint.
	// (POST /shaman/checkout/create)
	ShamanCheckout(ctx echo.Context) error
	// Checks a Shaman Requirements file, and reports which files are unknown.
	// (POST /shaman/checkout/requirements)
	ShamanCheckoutRequirements(ctx echo.Context) error
	// Check the status of a file on the Shaman server.
	// (GET /shaman/files/{checksum}/{filesize})
	ShamanFileStoreCheck(ctx echo.Context, checksum string, filesize int) error
	// Store a new file on the Shaman server. Note that the Shaman server can forcibly close the HTTP connection when another client finishes uploading the exact same file, to prevent double uploads.
	// The file's contents should be sent in the request body.
	// (POST /shaman/files/{checksum}/{filesize})
	ShamanFileStore(ctx echo.Context, checksum string, filesize int, params ShamanFileStoreParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetConfiguration converts echo context to params.
func (w *ServerInterfaceWrapper) GetConfiguration(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetConfiguration(ctx)
	return err
}

// SubmitJob converts echo context to params.
func (w *ServerInterfaceWrapper) SubmitJob(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SubmitJob(ctx)
	return err
}

// FetchGlobalLastRenderedInfo converts echo context to params.
func (w *ServerInterfaceWrapper) FetchGlobalLastRenderedInfo(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchGlobalLastRenderedInfo(ctx)
	return err
}

// QueryJobs converts echo context to params.
func (w *ServerInterfaceWrapper) QueryJobs(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.QueryJobs(ctx)
	return err
}

// GetJobType converts echo context to params.
func (w *ServerInterfaceWrapper) GetJobType(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "typeName" -------------
	var typeName string

	err = runtime.BindStyledParameterWithLocation("simple", false, "typeName", runtime.ParamLocationPath, ctx.Param("typeName"), &typeName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter typeName: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetJobType(ctx, typeName)
	return err
}

// GetJobTypes converts echo context to params.
func (w *ServerInterfaceWrapper) GetJobTypes(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetJobTypes(ctx)
	return err
}

// FetchJob converts echo context to params.
func (w *ServerInterfaceWrapper) FetchJob(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchJob(ctx, jobId)
	return err
}

// RemoveJobBlocklist converts echo context to params.
func (w *ServerInterfaceWrapper) RemoveJobBlocklist(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RemoveJobBlocklist(ctx, jobId)
	return err
}

// FetchJobBlocklist converts echo context to params.
func (w *ServerInterfaceWrapper) FetchJobBlocklist(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchJobBlocklist(ctx, jobId)
	return err
}

// FetchJobLastRenderedInfo converts echo context to params.
func (w *ServerInterfaceWrapper) FetchJobLastRenderedInfo(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchJobLastRenderedInfo(ctx, jobId)
	return err
}

// SetJobStatus converts echo context to params.
func (w *ServerInterfaceWrapper) SetJobStatus(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SetJobStatus(ctx, jobId)
	return err
}

// FetchJobTasks converts echo context to params.
func (w *ServerInterfaceWrapper) FetchJobTasks(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "job_id" -------------
	var jobId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "job_id", runtime.ParamLocationPath, ctx.Param("job_id"), &jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter job_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchJobTasks(ctx, jobId)
	return err
}

// FetchTask converts echo context to params.
func (w *ServerInterfaceWrapper) FetchTask(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchTask(ctx, taskId)
	return err
}

// FetchTaskLogTail converts echo context to params.
func (w *ServerInterfaceWrapper) FetchTaskLogTail(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchTaskLogTail(ctx, taskId)
	return err
}

// SetTaskStatus converts echo context to params.
func (w *ServerInterfaceWrapper) SetTaskStatus(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SetTaskStatus(ctx, taskId)
	return err
}

// GetVersion converts echo context to params.
func (w *ServerInterfaceWrapper) GetVersion(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetVersion(ctx)
	return err
}

// FetchWorkers converts echo context to params.
func (w *ServerInterfaceWrapper) FetchWorkers(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchWorkers(ctx)
	return err
}

// FetchWorker converts echo context to params.
func (w *ServerInterfaceWrapper) FetchWorker(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "worker_id" -------------
	var workerId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "worker_id", runtime.ParamLocationPath, ctx.Param("worker_id"), &workerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter worker_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.FetchWorker(ctx, workerId)
	return err
}

// RequestWorkerStatusChange converts echo context to params.
func (w *ServerInterfaceWrapper) RequestWorkerStatusChange(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "worker_id" -------------
	var workerId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "worker_id", runtime.ParamLocationPath, ctx.Param("worker_id"), &workerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter worker_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RequestWorkerStatusChange(ctx, workerId)
	return err
}

// RegisterWorker converts echo context to params.
func (w *ServerInterfaceWrapper) RegisterWorker(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RegisterWorker(ctx)
	return err
}

// SignOff converts echo context to params.
func (w *ServerInterfaceWrapper) SignOff(ctx echo.Context) error {
	var err error

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SignOff(ctx)
	return err
}

// SignOn converts echo context to params.
func (w *ServerInterfaceWrapper) SignOn(ctx echo.Context) error {
	var err error

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SignOn(ctx)
	return err
}

// WorkerState converts echo context to params.
func (w *ServerInterfaceWrapper) WorkerState(ctx echo.Context) error {
	var err error

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.WorkerState(ctx)
	return err
}

// WorkerStateChanged converts echo context to params.
func (w *ServerInterfaceWrapper) WorkerStateChanged(ctx echo.Context) error {
	var err error

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.WorkerStateChanged(ctx)
	return err
}

// ScheduleTask converts echo context to params.
func (w *ServerInterfaceWrapper) ScheduleTask(ctx echo.Context) error {
	var err error

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ScheduleTask(ctx)
	return err
}

// TaskUpdate converts echo context to params.
func (w *ServerInterfaceWrapper) TaskUpdate(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.TaskUpdate(ctx, taskId)
	return err
}

// MayWorkerRun converts echo context to params.
func (w *ServerInterfaceWrapper) MayWorkerRun(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.MayWorkerRun(ctx, taskId)
	return err
}

// TaskOutputProduced converts echo context to params.
func (w *ServerInterfaceWrapper) TaskOutputProduced(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "task_id" -------------
	var taskId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "task_id", runtime.ParamLocationPath, ctx.Param("task_id"), &taskId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter task_id: %s", err))
	}

	ctx.Set(Worker_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.TaskOutputProduced(ctx, taskId)
	return err
}

// ShamanCheckout converts echo context to params.
func (w *ServerInterfaceWrapper) ShamanCheckout(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShamanCheckout(ctx)
	return err
}

// ShamanCheckoutRequirements converts echo context to params.
func (w *ServerInterfaceWrapper) ShamanCheckoutRequirements(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShamanCheckoutRequirements(ctx)
	return err
}

// ShamanFileStoreCheck converts echo context to params.
func (w *ServerInterfaceWrapper) ShamanFileStoreCheck(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "checksum" -------------
	var checksum string

	err = runtime.BindStyledParameterWithLocation("simple", false, "checksum", runtime.ParamLocationPath, ctx.Param("checksum"), &checksum)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter checksum: %s", err))
	}

	// ------------- Path parameter "filesize" -------------
	var filesize int

	err = runtime.BindStyledParameterWithLocation("simple", false, "filesize", runtime.ParamLocationPath, ctx.Param("filesize"), &filesize)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter filesize: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShamanFileStoreCheck(ctx, checksum, filesize)
	return err
}

// ShamanFileStore converts echo context to params.
func (w *ServerInterfaceWrapper) ShamanFileStore(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "checksum" -------------
	var checksum string

	err = runtime.BindStyledParameterWithLocation("simple", false, "checksum", runtime.ParamLocationPath, ctx.Param("checksum"), &checksum)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter checksum: %s", err))
	}

	// ------------- Path parameter "filesize" -------------
	var filesize int

	err = runtime.BindStyledParameterWithLocation("simple", false, "filesize", runtime.ParamLocationPath, ctx.Param("filesize"), &filesize)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter filesize: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params ShamanFileStoreParams

	headers := ctx.Request().Header
	// ------------- Optional header parameter "X-Shaman-Can-Defer-Upload" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Shaman-Can-Defer-Upload")]; found {
		var XShamanCanDeferUpload bool
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Shaman-Can-Defer-Upload, got %d", n))
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "X-Shaman-Can-Defer-Upload", runtime.ParamLocationHeader, valueList[0], &XShamanCanDeferUpload)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Shaman-Can-Defer-Upload: %s", err))
		}

		params.XShamanCanDeferUpload = &XShamanCanDeferUpload
	}
	// ------------- Optional header parameter "X-Shaman-Original-Filename" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Shaman-Original-Filename")]; found {
		var XShamanOriginalFilename string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Shaman-Original-Filename, got %d", n))
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "X-Shaman-Original-Filename", runtime.ParamLocationHeader, valueList[0], &XShamanOriginalFilename)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Shaman-Original-Filename: %s", err))
		}

		params.XShamanOriginalFilename = &XShamanOriginalFilename
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShamanFileStore(ctx, checksum, filesize, params)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/api/configuration", wrapper.GetConfiguration)
	router.POST(baseURL+"/api/jobs", wrapper.SubmitJob)
	router.GET(baseURL+"/api/jobs/last-rendered", wrapper.FetchGlobalLastRenderedInfo)
	router.POST(baseURL+"/api/jobs/query", wrapper.QueryJobs)
	router.GET(baseURL+"/api/jobs/type/:typeName", wrapper.GetJobType)
	router.GET(baseURL+"/api/jobs/types", wrapper.GetJobTypes)
	router.GET(baseURL+"/api/jobs/:job_id", wrapper.FetchJob)
	router.DELETE(baseURL+"/api/jobs/:job_id/blocklist", wrapper.RemoveJobBlocklist)
	router.GET(baseURL+"/api/jobs/:job_id/blocklist", wrapper.FetchJobBlocklist)
	router.GET(baseURL+"/api/jobs/:job_id/last-rendered", wrapper.FetchJobLastRenderedInfo)
	router.POST(baseURL+"/api/jobs/:job_id/setstatus", wrapper.SetJobStatus)
	router.GET(baseURL+"/api/jobs/:job_id/tasks", wrapper.FetchJobTasks)
	router.GET(baseURL+"/api/tasks/:task_id", wrapper.FetchTask)
	router.GET(baseURL+"/api/tasks/:task_id/logtail", wrapper.FetchTaskLogTail)
	router.POST(baseURL+"/api/tasks/:task_id/setstatus", wrapper.SetTaskStatus)
	router.GET(baseURL+"/api/version", wrapper.GetVersion)
	router.GET(baseURL+"/api/worker-mgt/workers", wrapper.FetchWorkers)
	router.GET(baseURL+"/api/worker-mgt/workers/:worker_id", wrapper.FetchWorker)
	router.POST(baseURL+"/api/worker-mgt/workers/:worker_id/setstatus", wrapper.RequestWorkerStatusChange)
	router.POST(baseURL+"/api/worker/register-worker", wrapper.RegisterWorker)
	router.POST(baseURL+"/api/worker/sign-off", wrapper.SignOff)
	router.POST(baseURL+"/api/worker/sign-on", wrapper.SignOn)
	router.GET(baseURL+"/api/worker/state", wrapper.WorkerState)
	router.POST(baseURL+"/api/worker/state-changed", wrapper.WorkerStateChanged)
	router.POST(baseURL+"/api/worker/task", wrapper.ScheduleTask)
	router.POST(baseURL+"/api/worker/task/:task_id", wrapper.TaskUpdate)
	router.GET(baseURL+"/api/worker/task/:task_id/may-i-run", wrapper.MayWorkerRun)
	router.POST(baseURL+"/api/worker/task/:task_id/output-produced", wrapper.TaskOutputProduced)
	router.POST(baseURL+"/shaman/checkout/create", wrapper.ShamanCheckout)
	router.POST(baseURL+"/shaman/checkout/requirements", wrapper.ShamanCheckoutRequirements)
	router.GET(baseURL+"/shaman/files/:checksum/:filesize", wrapper.ShamanFileStoreCheck)
	router.POST(baseURL+"/shaman/files/:checksum/:filesize", wrapper.ShamanFileStore)

}
