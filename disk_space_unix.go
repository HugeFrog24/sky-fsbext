//go:build !windows
// +build !windows

package main

import (
	"log"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func CheckDiskSpace(outputDir string, requiredSpace uint64) {
	var stat unix.Statfs_t
	err := unix.Statfs(filepath.Dir(outputDir), &stat)
	if err != nil {
		log.Fatalf("Failed to get disk space: %v\n", err)
	}

	// Available blocks * size per block = available space in bytes
	availableSpace := stat.Bavail * uint64(stat.Bsize)
	if availableSpace < requiredSpace {
		log.Fatalf("Insufficient disk space. Required: %d bytes, Available: %d bytes\n", requiredSpace, availableSpace)
	}
}
