// Code generated by mockery v1.0.0. DO NOT EDIT.

package channel

import mock "github.com/stretchr/testify/mock"
import time "time"

// mockWsConnInterface is an autogenerated mock type for the wsConnInterface type
type mockWsConnInterface struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *mockWsConnInterface) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReadJSON provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) ReadJSON(_a0 interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPongHandler provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) SetPongHandler(_a0 func(string) error) {
	_m.Called(_a0)
}

// SetReadDeadline provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) SetReadDeadline(_a0 time.Time) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Time) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetReadLimit provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) SetReadLimit(_a0 int64) {
	_m.Called(_a0)
}

// SetWriteDeadline provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) SetWriteDeadline(_a0 time.Time) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Time) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteJSON provides a mock function with given fields: _a0
func (_m *mockWsConnInterface) WriteJSON(_a0 interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteMessage provides a mock function with given fields: _a0, _a1
func (_m *mockWsConnInterface) WriteMessage(_a0 int, _a1 []byte) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, []byte) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}