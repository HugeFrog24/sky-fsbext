//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// findSteamGamePath attempts to find the Steam installation path and returns the game path for Sky
func findSteamGamePath() (string, error) {
	steamPath, err := findSteamInstallPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(steamPath, "steamapps", "common", "Sky Children of the Light"), nil
}

// findSteamInstallPath attempts to find the Steam installation path from the Windows registry
func findSteamInstallPath() (string, error) {
	// Attempt to open 32-bit registry key
	keyPath32 := `SOFTWARE\Valve\Steam`
	key32, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath32, registry.QUERY_VALUE)
	if err == nil {
		defer key32.Close()
		installPath, _, err := key32.GetStringValue("InstallPath")
		if err == nil && installPath != "" {
			return installPath, nil
		}
	}

	// Attempt to open 64-bit registry key
	keyPath64 := `SOFTWARE\Wow6432Node\Valve\Steam`
	key64, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath64, registry.QUERY_VALUE)
	if err == nil {
		defer key64.Close()
		installPath, _, err := key64.GetStringValue("InstallPath")
		if err == nil && installPath != "" {
			return installPath, nil
		}
	}

	return "", fmt.Errorf("steam InstallPath not found in registry")
}

// getSkyAudioPaths returns all paths to Sky's audio files within the Steam installation
func getSkyAudioPaths() ([]string, error) {
	skyPath, err := findSteamGamePath()
	if err != nil {
		return nil, err
	}

	// Base path for audio assets
	assetsPath := filepath.Join(skyPath, "data", "assets")
	if !dirExists(assetsPath) {
		return nil, fmt.Errorf("Sky assets directory not found at %s", assetsPath)
	}

	var audioPaths []string

	// Read all subdirectories in assets
	entries, err := os.ReadDir(assetsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assets directory: %v", err)
	}

	// Check each subdirectory for the audio pattern
	for _, entry := range entries {
		if entry.IsDir() {
			audioPath := filepath.Join(assetsPath, entry.Name(), "Data", "Audio", "Fmod", "fmodswitch")
			if dirExists(audioPath) {
				audioPaths = append(audioPaths, audioPath)
			}
		}
	}

	if len(audioPaths) == 0 {
		return nil, fmt.Errorf("no Sky audio directories found in Steam installation at %s", skyPath)
	}

	return audioPaths, nil
}

// getSkyAudioPath returns the first available Sky audio path (for backward compatibility)
func getSkyAudioPath() (string, error) {
	paths, err := getSkyAudioPaths()
	if err != nil {
		return "", err
	}
	return paths[0], nil
}

// getSteamBankFiles returns all .bank files found in Steam installation
func getSteamBankFiles() ([]string, error) {
	audioPaths, err := getSkyAudioPaths()
	if err != nil {
		return nil, err
	}

	var allBankFiles []string
	for _, audioPath := range audioPaths {
		bankFiles, err := filepath.Glob(filepath.Join(audioPath, "*.bank"))
		if err != nil {
			log.Printf("Failed to search for .bank files in %s: %v", audioPath, err)
			continue
		}
		allBankFiles = append(allBankFiles, bankFiles...)
	}

	return allBankFiles, nil
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
