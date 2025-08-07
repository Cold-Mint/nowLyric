package lyrics

/*
#cgo LDFLAGS: -lrt
#include <string.h>   // 让 cgo 找到 memcpy
*/
import "C"
import (
	"path/filepath"
	"strings"
	"unsafe"
)

const Size = 1024

func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg", ".opus":
		return true
	default:
		return false
	}
}

// WriteCString 将 Go 字符串 s 写入 ptr 指向的共享内存（带 '\0' 结尾）
// Write the Go string 's' to the shared memory pointed to by ptr (ending with '\0')
func WriteCString(ptr unsafe.Pointer, s string, maxLen int) {
	if len(s) >= maxLen {
		s = s[:maxLen-1] // 预留 '\0'
	}
	C.memcpy(ptr, unsafe.Pointer(C.CString(s)), C.size_t(len(s)+1))
}
