// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/blender/flamenco-ng-poc/internal/manager/api_impl (interfaces: PersistenceService,JobCompiler)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	job_compilers "gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	persistence "gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	api "gitlab.com/blender/flamenco-ng-poc/pkg/api"
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

// CreateWorker mocks base method.
func (m *MockPersistenceService) CreateWorker(arg0 context.Context, arg1 *persistence.Worker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWorker", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateWorker indicates an expected call of CreateWorker.
func (mr *MockPersistenceServiceMockRecorder) CreateWorker(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWorker", reflect.TypeOf((*MockPersistenceService)(nil).CreateWorker), arg0, arg1)
}

// FetchJob mocks base method.
func (m *MockPersistenceService) FetchJob(arg0 context.Context, arg1 string) (*persistence.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchJob", arg0, arg1)
	ret0, _ := ret[0].(*persistence.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchJob indicates an expected call of FetchJob.
func (mr *MockPersistenceServiceMockRecorder) FetchJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchJob", reflect.TypeOf((*MockPersistenceService)(nil).FetchJob), arg0, arg1)
}

// FetchTask mocks base method.
func (m *MockPersistenceService) FetchTask(arg0 context.Context, arg1 string) (*persistence.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchTask", arg0, arg1)
	ret0, _ := ret[0].(*persistence.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchTask indicates an expected call of FetchTask.
func (mr *MockPersistenceServiceMockRecorder) FetchTask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchTask", reflect.TypeOf((*MockPersistenceService)(nil).FetchTask), arg0, arg1)
}

// FetchWorker mocks base method.
func (m *MockPersistenceService) FetchWorker(arg0 context.Context, arg1 string) (*persistence.Worker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchWorker", arg0, arg1)
	ret0, _ := ret[0].(*persistence.Worker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchWorker indicates an expected call of FetchWorker.
func (mr *MockPersistenceServiceMockRecorder) FetchWorker(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchWorker", reflect.TypeOf((*MockPersistenceService)(nil).FetchWorker), arg0, arg1)
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

// SaveWorker mocks base method.
func (m *MockPersistenceService) SaveWorker(arg0 context.Context, arg1 *persistence.Worker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveWorker", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveWorker indicates an expected call of SaveWorker.
func (mr *MockPersistenceServiceMockRecorder) SaveWorker(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveWorker", reflect.TypeOf((*MockPersistenceService)(nil).SaveWorker), arg0, arg1)
}

// SaveWorkerStatus mocks base method.
func (m *MockPersistenceService) SaveWorkerStatus(arg0 context.Context, arg1 *persistence.Worker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveWorkerStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveWorkerStatus indicates an expected call of SaveWorkerStatus.
func (mr *MockPersistenceServiceMockRecorder) SaveWorkerStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveWorkerStatus", reflect.TypeOf((*MockPersistenceService)(nil).SaveWorkerStatus), arg0, arg1)
}

// ScheduleTask mocks base method.
func (m *MockPersistenceService) ScheduleTask(arg0 *persistence.Worker) (*persistence.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScheduleTask", arg0)
	ret0, _ := ret[0].(*persistence.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ScheduleTask indicates an expected call of ScheduleTask.
func (mr *MockPersistenceServiceMockRecorder) ScheduleTask(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScheduleTask", reflect.TypeOf((*MockPersistenceService)(nil).ScheduleTask), arg0)
}

// StoreAuthoredJob mocks base method.
func (m *MockPersistenceService) StoreAuthoredJob(arg0 context.Context, arg1 job_compilers.AuthoredJob) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreAuthoredJob", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreAuthoredJob indicates an expected call of StoreAuthoredJob.
func (mr *MockPersistenceServiceMockRecorder) StoreAuthoredJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreAuthoredJob", reflect.TypeOf((*MockPersistenceService)(nil).StoreAuthoredJob), arg0, arg1)
}

// MockJobCompiler is a mock of JobCompiler interface.
type MockJobCompiler struct {
	ctrl     *gomock.Controller
	recorder *MockJobCompilerMockRecorder
}

// MockJobCompilerMockRecorder is the mock recorder for MockJobCompiler.
type MockJobCompilerMockRecorder struct {
	mock *MockJobCompiler
}

// NewMockJobCompiler creates a new mock instance.
func NewMockJobCompiler(ctrl *gomock.Controller) *MockJobCompiler {
	mock := &MockJobCompiler{ctrl: ctrl}
	mock.recorder = &MockJobCompilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobCompiler) EXPECT() *MockJobCompilerMockRecorder {
	return m.recorder
}

// Compile mocks base method.
func (m *MockJobCompiler) Compile(arg0 context.Context, arg1 api.SubmittedJob) (*job_compilers.AuthoredJob, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Compile", arg0, arg1)
	ret0, _ := ret[0].(*job_compilers.AuthoredJob)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Compile indicates an expected call of Compile.
func (mr *MockJobCompilerMockRecorder) Compile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Compile", reflect.TypeOf((*MockJobCompiler)(nil).Compile), arg0, arg1)
}

// ListJobTypes mocks base method.
func (m *MockJobCompiler) ListJobTypes() api.AvailableJobTypes {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobTypes")
	ret0, _ := ret[0].(api.AvailableJobTypes)
	return ret0
}

// ListJobTypes indicates an expected call of ListJobTypes.
func (mr *MockJobCompilerMockRecorder) ListJobTypes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobTypes", reflect.TypeOf((*MockJobCompiler)(nil).ListJobTypes))
}
