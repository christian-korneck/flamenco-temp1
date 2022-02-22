// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/blender/flamenco-ng-poc/internal/worker (interfaces: CommandLineRunner)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	exec "os/exec"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCommandLineRunner is a mock of CommandLineRunner interface.
type MockCommandLineRunner struct {
	ctrl     *gomock.Controller
	recorder *MockCommandLineRunnerMockRecorder
}

// MockCommandLineRunnerMockRecorder is the mock recorder for MockCommandLineRunner.
type MockCommandLineRunnerMockRecorder struct {
	mock *MockCommandLineRunner
}

// NewMockCommandLineRunner creates a new mock instance.
func NewMockCommandLineRunner(ctrl *gomock.Controller) *MockCommandLineRunner {
	mock := &MockCommandLineRunner{ctrl: ctrl}
	mock.recorder = &MockCommandLineRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommandLineRunner) EXPECT() *MockCommandLineRunnerMockRecorder {
	return m.recorder
}

// CommandContext mocks base method.
func (m *MockCommandLineRunner) CommandContext(arg0 context.Context, arg1 string, arg2 ...string) *exec.Cmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CommandContext", varargs...)
	ret0, _ := ret[0].(*exec.Cmd)
	return ret0
}

// CommandContext indicates an expected call of CommandContext.
func (mr *MockCommandLineRunnerMockRecorder) CommandContext(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommandContext", reflect.TypeOf((*MockCommandLineRunner)(nil).CommandContext), varargs...)
}