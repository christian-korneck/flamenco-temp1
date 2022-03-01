// Code generated by MockGen. DO NOT EDIT.
// Source: git.blender.org/flamenco/internal/manager/task_state_machine (interfaces: PersistenceService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	persistence "git.blender.org/flamenco/internal/manager/persistence"
	api "git.blender.org/flamenco/pkg/api"
	gomock "github.com/golang/mock/gomock"
)

// MockPersistenceService is a mock of PersistenceService interface.
type MockPersistenceService struct {
	ctrl     *gomock.Controller
	recorder *MockPersistenceServiceMockRecorder
}

// MockPersistenceServiceMockRecorder is the mock recorder for MockPersistenceService.
type MockPersistenceServiceMockRecorder struct {
	mock *MockPersistenceService
}

// NewMockPersistenceService creates a new mock instance.
func NewMockPersistenceService(ctrl *gomock.Controller) *MockPersistenceService {
	mock := &MockPersistenceService{ctrl: ctrl}
	mock.recorder = &MockPersistenceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPersistenceService) EXPECT() *MockPersistenceServiceMockRecorder {
	return m.recorder
}

// CountTasksOfJobInStatus mocks base method.
func (m *MockPersistenceService) CountTasksOfJobInStatus(arg0 context.Context, arg1 *persistence.Job, arg2 api.TaskStatus) (int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountTasksOfJobInStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CountTasksOfJobInStatus indicates an expected call of CountTasksOfJobInStatus.
func (mr *MockPersistenceServiceMockRecorder) CountTasksOfJobInStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountTasksOfJobInStatus", reflect.TypeOf((*MockPersistenceService)(nil).CountTasksOfJobInStatus), arg0, arg1, arg2)
}

// JobHasTasksInStatus mocks base method.
func (m *MockPersistenceService) JobHasTasksInStatus(arg0 context.Context, arg1 *persistence.Job, arg2 api.TaskStatus) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JobHasTasksInStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// JobHasTasksInStatus indicates an expected call of JobHasTasksInStatus.
func (mr *MockPersistenceServiceMockRecorder) JobHasTasksInStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JobHasTasksInStatus", reflect.TypeOf((*MockPersistenceService)(nil).JobHasTasksInStatus), arg0, arg1, arg2)
}

// SaveJobStatus mocks base method.
func (m *MockPersistenceService) SaveJobStatus(arg0 context.Context, arg1 *persistence.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveJobStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveJobStatus indicates an expected call of SaveJobStatus.
func (mr *MockPersistenceServiceMockRecorder) SaveJobStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveJobStatus", reflect.TypeOf((*MockPersistenceService)(nil).SaveJobStatus), arg0, arg1)
}

// SaveTask mocks base method.
func (m *MockPersistenceService) SaveTask(arg0 context.Context, arg1 *persistence.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveTask", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveTask indicates an expected call of SaveTask.
func (mr *MockPersistenceServiceMockRecorder) SaveTask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveTask", reflect.TypeOf((*MockPersistenceService)(nil).SaveTask), arg0, arg1)
}

// UpdateJobsTaskStatuses mocks base method.
func (m *MockPersistenceService) UpdateJobsTaskStatuses(arg0 context.Context, arg1 *persistence.Job, arg2 api.TaskStatus, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobsTaskStatuses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobsTaskStatuses indicates an expected call of UpdateJobsTaskStatuses.
func (mr *MockPersistenceServiceMockRecorder) UpdateJobsTaskStatuses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobsTaskStatuses", reflect.TypeOf((*MockPersistenceService)(nil).UpdateJobsTaskStatuses), arg0, arg1, arg2, arg3)
}

// UpdateJobsTaskStatusesConditional mocks base method.
func (m *MockPersistenceService) UpdateJobsTaskStatusesConditional(arg0 context.Context, arg1 *persistence.Job, arg2 []api.TaskStatus, arg3 api.TaskStatus, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobsTaskStatusesConditional", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobsTaskStatusesConditional indicates an expected call of UpdateJobsTaskStatusesConditional.
func (mr *MockPersistenceServiceMockRecorder) UpdateJobsTaskStatusesConditional(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobsTaskStatusesConditional", reflect.TypeOf((*MockPersistenceService)(nil).UpdateJobsTaskStatusesConditional), arg0, arg1, arg2, arg3, arg4)
}
