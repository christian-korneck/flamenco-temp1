package task_state_machine

import (
	"testing"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRequeueActiveTasksOfWorker(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	worker := persistence.Worker{
		UUID: "3ed470c8-d41e-4668-92d0-d799997433a4",
		Name: "testert",
	}

	// Mock that the worker has two active tasks. It shouldn't happen, but even
	// when it does, both should be requeued when the worker signs off.
	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusActive)
	task2 := taskOfSameJob(task1, api.TaskStatusActive)
	workerTasks := []*persistence.Task{task1, task2}

	task1PrevStatus := task1.Status
	task2PrevStatus := task2.Status

	mocks.persist.EXPECT().FetchTasksOfWorkerInStatus(ctx, &worker, api.TaskStatusActive).Return(workerTasks, nil)

	// Expect this re-queueing to end up in the task's log and activity.
	mocks.persist.EXPECT().SaveTaskActivity(ctx, task1) // TODO: test saved activity value
	mocks.persist.EXPECT().SaveTaskActivity(ctx, task2) // TODO: test saved activity value
	mocks.persist.EXPECT().SaveTask(ctx, task1)         // TODO: test saved task status
	mocks.persist.EXPECT().SaveTask(ctx, task2)         // TODO: test saved task status

	logMsg := "Task was requeued by Manager because worker had to test"
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), task1.Job.UUID, task1.UUID, logMsg)
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), task2.Job.UUID, task2.UUID, logMsg)

	mocks.broadcaster.EXPECT().BroadcastTaskUpdate(api.SocketIOTaskUpdate{
		Activity:       logMsg,
		Id:             task1.UUID,
		JobId:          task1.Job.UUID,
		Name:           task1.Name,
		PreviousStatus: &task1PrevStatus,
		Status:         api.TaskStatusQueued,
		Updated:        task1.UpdatedAt,
	})

	mocks.broadcaster.EXPECT().BroadcastTaskUpdate(api.SocketIOTaskUpdate{
		Activity:       logMsg,
		Id:             task2.UUID,
		JobId:          task2.Job.UUID,
		Name:           task2.Name,
		PreviousStatus: &task2PrevStatus,
		Status:         api.TaskStatusQueued,
		Updated:        task2.UpdatedAt,
	})

	err := sm.RequeueActiveTasksOfWorker(ctx, &worker, "worker had to test")
	assert.NoError(t, err)
}
