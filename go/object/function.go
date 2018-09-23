package object

import (
	"runtime"
	"reflect"
	"path/filepath"
	"strings"
)

func GetFunctionFullName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
func GetFunctionName(i interface{}) string {
	nameFull := GetFunctionFullName(i)
	nameEnd := filepath.Ext(nameFull)        // .foo-fm
	name := strings.TrimPrefix(nameEnd, ".") // foo-fm
	name = strings.Split(name, "-")[0]       // foo
	return name
}
