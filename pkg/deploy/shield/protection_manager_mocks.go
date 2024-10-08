// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/aws-load-balancer-controller/pkg/deploy/shield (interfaces: ProtectionManager)

// Package shield is a generated GoMock package.
package shield

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockProtectionManager is a mock of ProtectionManager interface.
type MockProtectionManager struct {
	ctrl     *gomock.Controller
	recorder *MockProtectionManagerMockRecorder
}

// MockProtectionManagerMockRecorder is the mock recorder for MockProtectionManager.
type MockProtectionManagerMockRecorder struct {
	mock *MockProtectionManager
}

// NewMockProtectionManager creates a new mock instance.
func NewMockProtectionManager(ctrl *gomock.Controller) *MockProtectionManager {
	mock := &MockProtectionManager{ctrl: ctrl}
	mock.recorder = &MockProtectionManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProtectionManager) EXPECT() *MockProtectionManagerMockRecorder {
	return m.recorder
}

// CreateProtection mocks base method.
func (m *MockProtectionManager) CreateProtection(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProtection", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateProtection indicates an expected call of CreateProtection.
func (mr *MockProtectionManagerMockRecorder) CreateProtection(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProtection", reflect.TypeOf((*MockProtectionManager)(nil).CreateProtection), arg0, arg1, arg2)
}

// DeleteProtection mocks base method.
func (m *MockProtectionManager) DeleteProtection(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProtection", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProtection indicates an expected call of DeleteProtection.
func (mr *MockProtectionManagerMockRecorder) DeleteProtection(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProtection", reflect.TypeOf((*MockProtectionManager)(nil).DeleteProtection), arg0, arg1, arg2)
}

// GetProtection mocks base method.
func (m *MockProtectionManager) GetProtection(arg0 context.Context, arg1 string) (*ProtectionInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProtection", arg0, arg1)
	ret0, _ := ret[0].(*ProtectionInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProtection indicates an expected call of GetProtection.
func (mr *MockProtectionManagerMockRecorder) GetProtection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProtection", reflect.TypeOf((*MockProtectionManager)(nil).GetProtection), arg0, arg1)
}

// IsSubscribed mocks base method.
func (m *MockProtectionManager) IsSubscribed(arg0 context.Context) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSubscribed", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsSubscribed indicates an expected call of IsSubscribed.
func (mr *MockProtectionManagerMockRecorder) IsSubscribed(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSubscribed", reflect.TypeOf((*MockProtectionManager)(nil).IsSubscribed), arg0)
}
