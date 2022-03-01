// Code generated by MockGen. DO NOT EDIT.
// Source: git.blender.org/flamenco/internal/worker (interfaces: TaskExecutionListener)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTaskExecutionListener is a mock of TaskExecutionListener interface.
type MockTaskExecutionListener struct {
	ctrl     *gomock.Controller
	recorder *MockTaskExecutionListenerMockRecorder
}

// MockTaskExecutionListenerMockRecorder is the mock recorder for MockTaskExecutionListener.
type MockTaskExecutionListenerMockRecorder struct {
	mock *MockTaskExecutionListener
}

// NewMockTaskExecutionListener creates a new mock instance.
func NewMockTaskExecutionListener(ctrl *gomock.Controller) *MockTaskExecutionListener {
	mock := &MockTaskExecutionListener{ctrl: ctrl}
	mock.recorder = &MockTaskExecutionListenerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskExecutionListener) EXPECT() *MockTaskExecutionListenerMockRecorder {
	return m.recorder
}

// TaskCompleted mocks base method.
func (m *MockTaskExecutionListener) TaskCompleted(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskCompleted", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// TaskCompleted indicates an expected call of TaskCompleted.
func (mr *MockTaskExecutionListenerMockRecorder) TaskCompleted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskCompleted", reflect.TypeOf((*MockTaskExecutionListener)(nil).TaskCompleted), arg0, arg1)
}

// TaskFailed mocks base method.
func (m *MockTaskExecutionListener) TaskFailed(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskFailed", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// TaskFailed indicates an expected call of TaskFailed.
func (mr *MockTaskExecutionListenerMockRecorder) TaskFailed(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskFailed", reflect.TypeOf((*MockTaskExecutionListener)(nil).TaskFailed), arg0, arg1, arg2)
}

// TaskStarted mocks base method.
func (m *MockTaskExecutionListener) TaskStarted(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskStarted", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// TaskStarted indicates an expected call of TaskStarted.
func (mr *MockTaskExecutionListenerMockRecorder) TaskStarted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskStarted", reflect.TypeOf((*MockTaskExecutionListener)(nil).TaskStarted), arg0, arg1)
}
