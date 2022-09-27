package reqest

import (
	"net/http"
	"reflect"

	"github.com/golang/mock/gomock"
)

// MockRoundTripper is a mock of RoundTripper interface.
type MockRoundTripper struct {
	ctrl     *gomock.Controller
	recorder *MockRoundTripperMockRecorder
}

// MockRoundTripperMockRecorder is the mock recorder for MockBuildProcess.
type MockRoundTripperMockRecorder struct {
	mock *MockRoundTripper
}

// NewMockRoundTripper creates a new mock instance.
func NewMockRoundTripper(ctrl *gomock.Controller) *MockRoundTripper {
	mock := &MockRoundTripper{ctrl: ctrl}
	mock.recorder = &MockRoundTripperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoundTripper) EXPECT() *MockRoundTripperMockRecorder {
	return m.recorder
}

// RoundTrip mocks base method.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RoundTrip", req)
	return ret[0].(*http.Response), ret[1].(error)
}

// RoundTrip indicates an expected call of RoundTrip.
func (mr *MockRoundTripperMockRecorder) RoundTrip(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RoundTrip", reflect.TypeOf((*MockRoundTripper)(nil).RoundTrip), req)
}
