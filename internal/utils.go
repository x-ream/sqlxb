package internal

import (
	"reflect"
	"strings"
)

func IsOmit(column string, tag string, v interface{}) bool {

	t:=reflect.TypeOf(v)
	for i:=0;i<t.NumField();i++{
		f := t.Field(i)
		tag := f.Tag.Get(tag)
		if strings.Contains(tag, column) && strings.Contains(tag,"omitempty") {
			return true
		}
	}

	return false
}
