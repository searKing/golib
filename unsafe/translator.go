package unsafe

import "C"
import (
	"reflect"
	"unsafe"
)

// []string -> char**
func CStringArray(strs ...string) (**C.char, C.int) {
	// []string -> [](*C.char)
	cCharArray := make([]*C.char, 0, len(strs))
	for _, s := range strs {
		cCharArray = append(cCharArray, (*C.char)(unsafe.Pointer(C.CString(s))))
	}
	return (**C.char)(unsafe.Pointer(&cCharArray[0])), C.int(len(strs))
}

// char** -> []string
func GoStringArray(strArray unsafe.Pointer, n int) []string {
	// char** -> [](C.*char)
	cCharArray := make([]*C.char, n)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&cCharArray))
	header.Cap = n
	header.Len = n
	header.Data = uintptr(strArray)

	// [](C.*char) -> []string
	strs := make([]string, 0, n)
	for _, s := range cCharArray {
		strs = append(strs, C.GoString(s))
	}
	return strs
}
