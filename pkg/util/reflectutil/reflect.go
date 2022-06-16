package reflectutil

import "reflect"

// TypeOf returns the reflect.Type that represents the type T.
func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
