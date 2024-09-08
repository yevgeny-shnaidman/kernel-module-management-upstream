// Code generated by MockGen. DO NOT EDIT.
// Source: filesystem_helper.go
//
// Generated by this command:
//
//	mockgen -source=filesystem_helper.go -package=utils -destination=mock_filesystem_helper.go
//
// Package utils is a generated GoMock package.
package utils

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockFSHelper is a mock of FSHelper interface.
type MockFSHelper struct {
	ctrl     *gomock.Controller
	recorder *MockFSHelperMockRecorder
}

// MockFSHelperMockRecorder is the mock recorder for MockFSHelper.
type MockFSHelperMockRecorder struct {
	mock *MockFSHelper
}

// NewMockFSHelper creates a new mock instance.
func NewMockFSHelper(ctrl *gomock.Controller) *MockFSHelper {
	mock := &MockFSHelper{ctrl: ctrl}
	mock.recorder = &MockFSHelperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFSHelper) EXPECT() *MockFSHelperMockRecorder {
	return m.recorder
}

// RemoveSrcFilesFromDst mocks base method.
func (m *MockFSHelper) RemoveSrcFilesFromDst(srcDir, dstDir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveSrcFilesFromDst", srcDir, dstDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveSrcFilesFromDst indicates an expected call of RemoveSrcFilesFromDst.
func (mr *MockFSHelperMockRecorder) RemoveSrcFilesFromDst(srcDir, dstDir any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveSrcFilesFromDst", reflect.TypeOf((*MockFSHelper)(nil).RemoveSrcFilesFromDst), srcDir, dstDir)
}
