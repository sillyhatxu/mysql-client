package dbclient

import "reflect"

func isStruct(input interface{}) bool {
	inputT := reflect.TypeOf(input)
	return inputT.Kind() == reflect.Struct || isStructPtr(inputT)
}

func isSlice(input interface{}) bool {
	inputT := reflect.TypeOf(input)
	return inputT.Kind() == reflect.Slice
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
