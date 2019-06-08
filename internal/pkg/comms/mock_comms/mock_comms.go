// Code generated by MockGen. DO NOT EDIT.
// Source: internal/pkg/comms/server.go

// Package mock_comms is a generated GoMock package.
package mock_comms

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIEventServer is a mock of IEventServer interface
type MockIEventServer struct {
	ctrl     *gomock.Controller
	recorder *MockIEventServerMockRecorder
}

// MockIEventServerMockRecorder is the mock recorder for MockIEventServer
type MockIEventServerMockRecorder struct {
	mock *MockIEventServer
}

// NewMockIEventServer creates a new mock instance
func NewMockIEventServer(ctrl *gomock.Controller) *MockIEventServer {
	mock := &MockIEventServer{ctrl: ctrl}
	mock.recorder = &MockIEventServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIEventServer) EXPECT() *MockIEventServerMockRecorder {
	return m.recorder
}

// AddEventToSendQueue mocks base method
func (m *MockIEventServer) AddEventToSendQueue(data []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddEventToSendQueue", data)
}

// AddEventToSendQueue indicates an expected call of AddEventToSendQueue
func (mr *MockIEventServerMockRecorder) AddEventToSendQueue(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEventToSendQueue", reflect.TypeOf((*MockIEventServer)(nil).AddEventToSendQueue), data)
}

// InitializeEventSystem mocks base method
func (m *MockIEventServer) InitializeEventSystem() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InitializeEventSystem")
}

// InitializeEventSystem indicates an expected call of InitializeEventSystem
func (mr *MockIEventServerMockRecorder) InitializeEventSystem() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitializeEventSystem", reflect.TypeOf((*MockIEventServer)(nil).InitializeEventSystem))
}

// Close mocks base method
func (m *MockIEventServer) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockIEventServerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIEventServer)(nil).Close))
}