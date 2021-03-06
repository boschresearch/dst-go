// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	wallet "perun.network/go-perun/wallet"
)

// ROContractRegistry is an autogenerated mock type for the ROContractRegistry type
type ROContractRegistry struct {
	mock.Mock
}

// Adjudicator provides a mock function with given fields:
func (_m *ROContractRegistry) Adjudicator() wallet.Address {
	ret := _m.Called()

	var r0 wallet.Address
	if rf, ok := ret.Get(0).(func() wallet.Address); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet.Address)
		}
	}

	return r0
}

// Asset provides a mock function with given fields: symbol
func (_m *ROContractRegistry) Asset(symbol string) (wallet.Address, bool) {
	ret := _m.Called(symbol)

	var r0 wallet.Address
	if rf, ok := ret.Get(0).(func(string) wallet.Address); ok {
		r0 = rf(symbol)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet.Address)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(symbol)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// AssetETH provides a mock function with given fields:
func (_m *ROContractRegistry) AssetETH() wallet.Address {
	ret := _m.Called()

	var r0 wallet.Address
	if rf, ok := ret.Get(0).(func() wallet.Address); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet.Address)
		}
	}

	return r0
}

// Assets provides a mock function with given fields:
func (_m *ROContractRegistry) Assets() map[string]string {
	ret := _m.Called()

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// Symbol provides a mock function with given fields: asset
func (_m *ROContractRegistry) Symbol(asset wallet.Address) (string, bool) {
	ret := _m.Called(asset)

	var r0 string
	if rf, ok := ret.Get(0).(func(wallet.Address) string); ok {
		r0 = rf(asset)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(wallet.Address) bool); ok {
		r1 = rf(asset)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Token provides a mock function with given fields: symbol
func (_m *ROContractRegistry) Token(symbol string) (wallet.Address, bool) {
	ret := _m.Called(symbol)

	var r0 wallet.Address
	if rf, ok := ret.Get(0).(func(string) wallet.Address); ok {
		r0 = rf(symbol)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet.Address)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(symbol)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}
