package dbclient

import (
	"reflect"
	"strconv"
	"time"
)

func setupBool(input bool) string {
	return strconv.FormatBool(input)
}

func setupInt(input int) string {
	return strconv.Itoa(input)
}

func setupInt64(input int64) string {
	return strconv.FormatInt(input, 10)
}

func setupTime(input time.Duration) string {
	//make sure 1ms<=t<24h
	if input < time.Millisecond || input >= 24*time.Hour {
		return ""
	}
	return input.String()
}

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
