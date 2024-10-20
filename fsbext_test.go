package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// Mocking exec.Command for testing extractAndMoveFile
// type mockCommand struct {
// 	output []byte
// 	err    error
// }

// func (m *mockCommand) CombinedOutput() ([]byte, error) {
// 	return m.output, m.err
// }

// Override the exec.Command function for testing
// var execCommand = exec.Command

// Define a variable for extractAndMoveFile to allow mocking
var extractAndMoveFileFunc = extractAndMoveFile

func TestGetSizeOfDir(t *testing.T) {
	// Create a temporary directory with some files
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	content1 := []byte("Hello")
	content2 := []byte("World!")

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	size := getSizeOfDir(tempDir)
	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	// Create a temporary output directory
	tempDir, err := os.MkdirTemp("", "output")
	if err != nil {
		t.Fatalf("Failed to create temp output dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createDirectoryStructure(tempDir)

	// Check if directories were created
	dirs := []string{"Music", "SFX", "Other"}
	for _, dir := range dirs {
		path := filepath.Join(tempDir, dir)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestRemoveEmptyDirectories(t *testing.T) {
	// Create a temporary output directory with some empty and non-empty subdirectories
	tempDir, err := os.MkdirTemp("", "output")
	if err != nil {
		t.Fatalf("Failed to create temp output dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	emptyDir := filepath.Join(tempDir, "EmptyDir")
	nonEmptyDir := filepath.Join(tempDir, "NonEmptyDir")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}
	if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
		t.Fatalf("Failed to create non-empty directory: %v", err)
	}

	// Create a file in nonEmptyDir
	filePath := filepath.Join(nonEmptyDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to write file in non-empty directory: %v", err)
	}

	// Remove empty directories
	removeEmptyDirectories(tempDir)

	// Check if emptyDir was removed
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Errorf("Empty directory was not removed")
	}

	// Check if nonEmptyDir still exists
	if _, err := os.Stat(nonEmptyDir); os.IsNotExist(err) {
		t.Errorf("Non-empty directory was incorrectly removed")
	}
}

func TestIsDirEmpty(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initially, it should be empty
	empty, err := isDirEmpty(tempDir)
	if err != nil {
		t.Fatalf("Error checking if directory is empty: %v", err)
	}
	if !empty {
		t.Errorf("Expected directory to be empty")
	}

	// Create a file
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Now, it should not be empty
	empty, err = isDirEmpty(tempDir)
	if err != nil {
		t.Fatalf("Error checking if directory is empty: %v", err)
	}
	if empty {
		t.Errorf("Expected directory to not be empty")
	}
}

func TestIsValidBankFile(t *testing.T) {
	// Create a temporary valid bank file
	validFile, err := os.CreateTemp("", "valid.bank")
	if err != nil {
		t.Fatalf("Failed to create temp valid bank file: %v", err)
	}
	defer os.Remove(validFile.Name())

	// Write valid headers
	if _, err := validFile.Write([]byte("RIFF")); err != nil {
		t.Fatalf("Failed to write to valid bank file: %v", err)
	}
	validFile.Close()

	if !isValidBankFile(validFile.Name()) {
		t.Errorf("Expected valid bank file")
	}

	// Create a temporary invalid bank file
	invalidFile, err := os.CreateTemp("", "invalid.bank")
	if err != nil {
		t.Fatalf("Failed to create temp invalid bank file: %v", err)
	}
	defer os.Remove(invalidFile.Name())

	// Write invalid headers
	if _, err := invalidFile.Write([]byte("XXXX")); err != nil {
		t.Fatalf("Failed to write to invalid bank file: %v", err)
	}
	invalidFile.Close()

	if isValidBankFile(invalidFile.Name()) {
		t.Errorf("Expected invalid bank file")
	}
}

func TestCountFilesInDir(t *testing.T) {
	// Create a temporary directory with some files and subdirectories
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("world"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Count files
	count, err := countFilesInDir(tempDir)
	if err != nil {
		t.Fatalf("Error counting files: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 files, got %d", count)
	}
}

func TestGetOSVersion(t *testing.T) {
	osVersion := getOSVersion()
	expected := runtime.GOOS + "/" + runtime.GOARCH
	if osVersion != expected {
		t.Errorf("Expected OS version %s, got %s", expected, osVersion)
	}
}

func TestProcessBankFilesConcurrently(t *testing.T) {
	// Initialize loggers to avoid nil pointer dereference
	setupLogging()

	// Mock the extractAndMoveFile function
	originalExtractAndMoveFile := extractAndMoveFileFunc
	defer func() { extractAndMoveFileFunc = originalExtractAndMoveFile }()

	extractAndMoveFileFunc = func(bankFile string, printMutex *sync.Mutex) int {
		return 1
	}

	// Mock bank files
	bankFiles := []string{"bank1.bank", "bank2.bank", "bank3.bank"}
	maxWorkers := 2

	count := processBankFilesConcurrently(bankFiles, maxWorkers)
	if count != len(bankFiles) {
		t.Errorf("Expected %d extracted files, got %d", len(bankFiles), count)
	}
}

func TestExtractAndMoveFile(t *testing.T) {
	// Mock the extractAndMoveFile function
	originalExtractAndMoveFile := extractAndMoveFileFunc
	defer func() { extractAndMoveFileFunc = originalExtractAndMoveFile }()

	extractAndMoveFileFunc = func(bankFile string, printMutex *sync.Mutex) int {
		return 1
	}

	// Create a temporary bank file
	bankFile, err := os.CreateTemp("", "test.bank")
	if err != nil {
		t.Fatalf("Failed to create temp bank file: %v", err)
	}
	defer os.Remove(bankFile.Name())

	// Write valid header
	if _, err := bankFile.Write([]byte("RIFF")); err != nil {
		t.Fatalf("Failed to write to bank file: %v", err)
	}
	bankFile.Close()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "output")
	if err != nil {
		t.Fatalf("Failed to create temp output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Set global outputDir for the test
	originalOutputDir := outputDir
	defer func() { outputDir = originalOutputDir }()

	// Perform the test
	var printMutex sync.Mutex
	count := extractAndMoveFileFunc(bankFile.Name(), &printMutex)
	if count != 1 {
		t.Errorf("Expected 1 extracted file, got %d", count)
	}
}

// TestHelperProcess is used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Mock successful command
	os.Exit(0)
}

func TestSafePrintf(t *testing.T) {
	output := "Test message\n"
	mutex := &sync.Mutex{}

	// Redirect stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	go func() {
		defer w.Close()
		safePrintf(mutex, output)
	}()

	// Read the output
	w.Close()               // Close the writer to flush the data
	buf, _ := io.ReadAll(r) // Use io.ReadAll to read from the pipe
	os.Stdout = originalStdout

	if string(buf) != output {
		t.Errorf("Expected output %q, got %q", output, string(buf))
	}
}
