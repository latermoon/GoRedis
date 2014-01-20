package tool

import (
	"reflect"
	"testing"
)

func CheckErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func CheckType(t *testing.T, reply interface{}, kind reflect.Kind) {
	if reflect.TypeOf(reply).Kind() != kind {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
}
