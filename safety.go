package main

import (
	"reflect"
)

// isNil safely checks if an interface value is nil.
// In Go, an interface is nil only if both type and value are nil.
// A typed nil (e.g., (*Type)(nil)) stored in an interface is NOT nil.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	k := rv.Kind()
	if k == reflect.Ptr || k == reflect.Interface || k == reflect.Slice || k == reflect.Map || k == reflect.Chan || k == reflect.Func {
		return rv.IsNil()
	}
	return false
}

// safeUnwrapSourceAdapter safely unwraps a sourceAdapter, returning the underlying Source and true if successful.
// Returns nil, false if the value is not a sourceAdapter or if the adapter's source is nil.
func safeUnwrapSourceAdapter(src Source) (Source, bool) {
	if src == nil {
		return nil, false
	}
	sa, ok := src.(*sourceAdapter)
	if !ok {
		return nil, false
	}
	if isNil(sa.s) {
		return nil, false
	}
	return sa.s, true
}

// safeUnwrapTargetAdapter safely unwraps a targetAdapter, returning the underlying Target and true if successful.
// Returns nil, false if the value is not a targetAdapter or if the adapter's target is nil.
func safeUnwrapTargetAdapter(tgt Target) (Target, bool) {
	if tgt == nil {
		return nil, false
	}
	ta, ok := tgt.(*targetAdapter)
	if !ok {
		return nil, false
	}
	if isNil(ta.t) {
		return nil, false
	}
	return ta.t, true
}

// safeDerefPtr safely dereferences a pointer, returning the value and true if non-nil.
// Returns zero value and false if the pointer is nil.
func safeDerefPtr[T any](ptr *T) (T, bool) {
	if ptr == nil {
		var zero T
		return zero, false
	}
	return *ptr, true
}
