package unsafe

// #include <stdlib.h>
//enum {
//	sizeofPtr = sizeof(void*),
//};
import "C"
import (
	"reflect"
	"unsafe"
)

// You can write as this directly.
//	// []string -> [](*C.char)
// 	var strs []string
//	cCharArray := make([]*C.char, 0, len(strs))
//	for _, s := range strs {
//		char := (*C.char)(unsafe.Pointer(C.CString(s)))
//		cCharArray = append(cCharArray, char)
//		defer C.free(unsafe.Pointer(char)) //释放内存
//	}
//	var cstr **C.char
// 	var cstrSize C.int
//	cstr = (**C.char)(unsafe.Pointer(&cCharArray[0]))
//	cstrSize =  C.int(len(strs))
func CStringArray(strs ...string) (**C.char, C.int) {
	// []string -> [](*C.char)
	totalLen := len(strs) * C.sizeofPtr
	for _, s := range strs {
		totalLen += len(s)
	}
	cCharArrayBuf := C.malloc(C.size_t(totalLen))
	cCharArray := make([]*C.char, 0, len(strs))
	for _, s := range strs {
		cCharArray = append(cCharArray, (*C.char)(unsafe.Pointer(C.CString(s))))
	}
	//return (**C.char)(unsafe.Pointer(&cCharArray[0])), C.int(len(strs))
	return (**C.char)(unsafe.Pointer(cCharArrayBuf)), C.int(len(strs))
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
