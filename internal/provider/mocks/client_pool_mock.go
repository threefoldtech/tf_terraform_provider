// Code generated by MockGen. DO NOT EDIT.
// Source: internal/node/client_pool.go

// Package mock_client is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client "github.com/threefoldtech/terraform-provider-grid/internal/node"
	subi "github.com/threefoldtech/terraform-provider-grid/pkg/subi"
)

// MockNodeClientGetter is a mock of NodeClientGetter interface.
type MockNodeClientGetter struct {
	ctrl     *gomock.Controller
	recorder *MockNodeClientGetterMockRecorder
}

// MockNodeClientGetterMockRecorder is the mock recorder for MockNodeClientGetter.
type MockNodeClientGetterMockRecorder struct {
	mock *MockNodeClientGetter
}

// NewMockNodeClientGetter creates a new mock instance.
func NewMockNodeClientGetter(ctrl *gomock.Controller) *MockNodeClientGetter {
	mock := &MockNodeClientGetter{ctrl: ctrl}
	mock.recorder = &MockNodeClientGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodeClientGetter) EXPECT() *MockNodeClientGetterMockRecorder {
	return m.recorder
}

// GetNodeClient mocks base method.
func (m *MockNodeClientGetter) GetNodeClient(sub subi.SubstrateExt, nodeID uint32) (*client.NodeClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeClient", sub, nodeID)
	ret0, _ := ret[0].(*client.NodeClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodeClient indicates an expected call of GetNodeClient.
func (mr *MockNodeClientGetterMockRecorder) GetNodeClient(sub, nodeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeClient", reflect.TypeOf((*MockNodeClientGetter)(nil).GetNodeClient), sub, nodeID)
}