// Code generated by MockGen. DO NOT EDIT.
// Source: git.blender.org/flamenco/internal/manager/task_state_machine (interfaces: PersistenceService,ChangeBroadcaster)

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
func (m *MockPersistenceService) CountTasksOfJobInStatus(arg0 context.Context, arg1 *persistence.Job, arg2 ...api.TaskStatus) (int, int, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CountTasksOfJobInStatus", varargs...)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CountTasksOfJobInStatus indicates an expected call of CountTasksOfJobInStatus.
func (mr *MockPersistenceServiceMockRecorder) CountTasksOfJobInStatus(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountTasksOfJobInStatus", reflect.TypeOf((*MockPersistenceService)(nil).CountTasksOfJobInStatus), varargs...)
}

// FetchJobsInStatus mocks base method.
func (m *MockPersistenceService) FetchJobsInStatus(arg0 context.Context, arg1 ...api.JobStatus) ([]*persistence.Job, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchJobsInStatus", varargs...)
	ret0, _ := ret[0].([]*persistence.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchJobsInStatus indicates an expected call of FetchJobsInStatus.
func (mr *MockPersistenceServiceMockRecorder) FetchJobsInStatus(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchJobsInStatus", reflect.TypeOf((*MockPersistenceService)(nil).FetchJobsInStatus), varargs...)
}

// FetchTasksOfJob mocks base method.
func (m *MockPersistenceService) FetchTasksOfJob(arg0 context.Context, arg1 *persistence.Job) ([]*persistence.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchTasksOfJob", arg0, arg1)
	ret0, _ := ret[0].([]*persistence.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchTasksOfJob indicates an expected call of FetchTasksOfJob.
func (mr *MockPersistenceServiceMockRecorder) FetchTasksOfJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchTasksOfJob", reflect.TypeOf((*MockPersistenceService)(nil).FetchTasksOfJob), arg0, arg1)
}

// FetchTasksOfJobInStatus mocks base method.
func (m *MockPersistenceService) FetchTasksOfJobInStatus(arg0 context.Context, arg1 *persistence.Job, arg2 ...api.TaskStatus) ([]*persistence.Task, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchTasksOfJobInStatus", varargs...)
	ret0, _ := ret[0].([]*persistence.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchTasksOfJobInStatus indicates an expected call of FetchTasksOfJobInStatus.
func (mr *MockPersistenceServiceMockRecorder) FetchTasksOfJobInStatus(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchTasksOfJobInStatus", reflect.TypeOf((*MockPersistenceService)(nil).FetchTasksOfJobInStatus), varargs...)
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

// MockChangeBroadcaster is a mock of ChangeBroadcaster interface.
type MockChangeBroadcaster struct {
	ctrl     *gomock.Controller
	recorder *MockChangeBroadcasterMockRecorder
}

// MockChangeBroadcasterMockRecorder is the mock recorder for MockChangeBroadcaster.
type MockChangeBroadcasterMockRecorder struct {
	mock *MockChangeBroadcaster
}

// NewMockChangeBroadcaster creates a new mock instance.
func NewMockChangeBroadcaster(ctrl *gomock.Controller) *MockChangeBroadcaster {
	mock := &MockChangeBroadcaster{ctrl: ctrl}
	mock.recorder = &MockChangeBroadcasterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChangeBroadcaster) EXPECT() *MockChangeBroadcasterMockRecorder {
	return m.recorder
}

// BroadcastJobUpdate mocks base method.
func (m *MockChangeBroadcaster) BroadcastJobUpdate(arg0 api.JobUpdate) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BroadcastJobUpdate", arg0)
}

// BroadcastJobUpdate indicates an expected call of BroadcastJobUpdate.
func (mr *MockChangeBroadcasterMockRecorder) BroadcastJobUpdate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastJobUpdate", reflect.TypeOf((*MockChangeBroadcaster)(nil).BroadcastJobUpdate), arg0)
}

// BroadcastTaskUpdate mocks base method.
func (m *MockChangeBroadcaster) BroadcastTaskUpdate(arg0 api.SocketIOTaskUpdate) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BroadcastTaskUpdate", arg0)
}

// BroadcastTaskUpdate indicates an expected call of BroadcastTaskUpdate.
func (mr *MockChangeBroadcasterMockRecorder) BroadcastTaskUpdate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastTaskUpdate", reflect.TypeOf((*MockChangeBroadcaster)(nil).BroadcastTaskUpdate), arg0)
}
