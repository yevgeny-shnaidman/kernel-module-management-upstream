// Code generated by MockGen. DO NOT EDIT.
// Source: daemonset.go

// Package controllers is a generated GoMock package.
package controllers

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1beta1 "github.com/qbarrand/oot-operator/api/v1beta1"
	v1 "k8s.io/api/apps/v1"
)

// MockDaemonSetCreator is a mock of DaemonSetCreator interface.
type MockDaemonSetCreator struct {
	ctrl     *gomock.Controller
	recorder *MockDaemonSetCreatorMockRecorder
}

// MockDaemonSetCreatorMockRecorder is the mock recorder for MockDaemonSetCreator.
type MockDaemonSetCreatorMockRecorder struct {
	mock *MockDaemonSetCreator
}

// NewMockDaemonSetCreator creates a new mock instance.
func NewMockDaemonSetCreator(ctrl *gomock.Controller) *MockDaemonSetCreator {
	mock := &MockDaemonSetCreator{ctrl: ctrl}
	mock.recorder = &MockDaemonSetCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDaemonSetCreator) EXPECT() *MockDaemonSetCreatorMockRecorder {
	return m.recorder
}

// SetAsDesired mocks base method.
func (m *MockDaemonSetCreator) SetAsDesired(ds *v1.DaemonSet, image string, mod *v1beta1.Module, kernelVersion string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAsDesired", ds, image, mod, kernelVersion)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAsDesired indicates an expected call of SetAsDesired.
func (mr *MockDaemonSetCreatorMockRecorder) SetAsDesired(ds, image, mod, kernelVersion interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAsDesired", reflect.TypeOf((*MockDaemonSetCreator)(nil).SetAsDesired), ds, image, mod, kernelVersion)
}
