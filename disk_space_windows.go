//go:build windows
// +build windows

package main

import (
	"log"
	"syscall"
	"unsafe"
)

func CheckDiskSpace(outputDir string, requiredSpace uint64) {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64

	_, _, err := c.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(outputDir))),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		log.Fatalf("Failed to get disk space: %v\n", err)
	}

	if uint64(freeBytesAvailable) < requiredSpace {
		log.Printf("Less than %.2f GB of disk space available for extraction (%.2f GB free)\n",
			float64(requiredSpace)/(1024*1024*1024), float64(freeBytesAvailable)/(1024*1024*1024))
	}
}
