package cmd

/*
#cgo LDFLAGS: -lrt
#include <fcntl.h>
#include <sys/mman.h>
#include <unistd.h>
#include <stdlib.h>   // 让 cgo 找到 free
*/
import "C"
import (
	"fmt"
	"nowlyric/lyrics"
	"unsafe"

	"github.com/spf13/cobra"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read the lyrics that are playing.",
	Long:  `Read the lyrics that are playing.`,
	Run: func(cmd *cobra.Command, args []string) {
		cName := C.CString(shmNameCStr)
		defer C.free(unsafe.Pointer(cName))
		fd, err := C.shm_open(cName, C.O_RDONLY, 0666)
		if fd < 0 {
			panic(fmt.Sprintf("shm_open failed: %v", err))
		}
		defer C.close(fd)

		ptr, err := C.mmap(nil, lyrics.Size, C.PROT_READ, C.MAP_SHARED, fd, 0)
		if ptr == C.MAP_FAILED {
			panic(fmt.Sprintf("mmap failed: %v", err))
		}
		defer C.munmap(ptr, lyrics.Size)

		data := C.GoString((*C.char)(ptr))
		fmt.Println(data)
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
