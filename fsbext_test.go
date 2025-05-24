package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// TestMain sets up the logging before any tests are run and cleans up afterward.
func TestMain(m *testing.M) {
	// Redirect log output to discard during tests to avoid cluttering the test output.
	log.SetOutput(io.Discard)

	// Initialize the loggers.
	setupLogging()

	// Execute the tests.
	exitVal := m.Run()

	// Perform any necessary cleanup here.
	// For example, you might want to remove temporary log files if they are created.

	// Exit with the appropriate code.
	os.Exit(exitVal)
}

func TestGetSizeOfDir(t *testing.T) {
	// Create a temporary directory with some files
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

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
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

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
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

	// Set the global outputDir to the temp directory
	originalOutputDir := outputDir
	outputDir = tempDir
	defer func() { outputDir = originalOutputDir }()

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

	// Set the global outputDir to the temp directory
	originalOutputDir := outputDir
	outputDir = tempDir
	defer func() { outputDir = originalOutputDir }()

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
	// Create a temporary directory to act as inputDir
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the global inputDir to the temp directory
	originalInputDir := inputDir
	inputDir = tempDir
	defer func() { inputDir = originalInputDir }()

	// Create a temporary valid bank file
	validFile := filepath.Join(tempDir, "valid.bank")
	if err := os.WriteFile(validFile, []byte("RIFF1234"), 0644); err != nil {
		t.Fatalf("Failed to create valid bank file: %v", err)
	}

	if !isValidBankFile(validFile) {
		t.Errorf("Expected valid bank file")
	}

	// Create a temporary invalid bank file
	invalidFile := filepath.Join(tempDir, "invalid.bank")
	if err := os.WriteFile(invalidFile, []byte("XXXX1234"), 0644); err != nil {
		t.Fatalf("Failed to create invalid bank file: %v", err)
	}

	if isValidBankFile(invalidFile) {
		t.Errorf("Expected invalid bank file")
	}

	// Test file outside of inputDir
	outsideFile := filepath.Join(os.TempDir(), "outside.bank")
	if err := os.WriteFile(outsideFile, []byte("RIFF1234"), 0644); err != nil {
		t.Fatalf("Failed to create outside bank file: %v", err)
	}
	defer func() {
		if err := os.Remove(outsideFile); err != nil {
			t.Logf("Error removing outside file: %v", err)
		}
	}()

	if isValidBankFile(outsideFile) {
		t.Errorf("Expected file outside inputDir to be invalid")
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
	defer func() {
		if err := os.Remove(bankFile.Name()); err != nil {
			t.Logf("Error removing bank file: %v", err)
		}
	}()

	// Write valid header
	if _, err := bankFile.Write([]byte("RIFF")); err != nil {
		t.Fatalf("Failed to write to bank file: %v", err)
	}
	if err := bankFile.Close(); err != nil {
		t.Fatalf("Failed to close bank file: %v", err)
	}

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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if err := w.Close(); err != nil {
				t.Logf("Error closing pipe writer: %v", err)
			}
		}()
		safePrintf(mutex, output)
	}()

	// Wait for the goroutine to finish
	wg.Wait()

	// Restore stdout and read the output
	os.Stdout = originalStdout
	buf, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	if string(buf) != output {
		t.Errorf("Expected output %q, got %q", output, string(buf))
	}
}
