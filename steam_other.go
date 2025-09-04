//go:build !windows

package main

import (
	"fmt"
)

// findSteamGamePath returns an error on non-Windows platforms
func findSteamGamePath() (string, error) {
	return "", fmt.Errorf("steam auto-detection is only supported on Windows")
}

// getSteamBankFiles returns an error on non-Windows platforms
func getSteamBankFiles() ([]string, error) {
	return nil, fmt.Errorf("steam auto-detection is only supported on Windows")
}
