package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	author  = "Tibik"
	version = "1.0.6"
)

var (
	verbose          bool
	inputDir         string
	outputDir        string
	vgmstreamPath    string
	compressionRatio float64
)

var (
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	summaryLogger *log.Logger
)

func init() {
	flag.StringVar(&inputDir, "i", "in", "Path to the input directory.")
	flag.StringVar(&inputDir, "input-dir", "in", "Path to the input directory.")
	flag.StringVar(&outputDir, "o", "out", "Path to the output directory.")
	flag.StringVar(&outputDir, "output-dir", "out", "Path to the output directory.")
	flag.StringVar(&vgmstreamPath, "p", filepath.Join("vgmstream-win64", "vgmstream-cli.exe"), "Path to vgmstream-cli executable.")
	flag.StringVar(&vgmstreamPath, "vgmstream-path", filepath.Join("vgmstream-win64", "vgmstream-cli.exe"), "Path to vgmstream-cli executable.")
	flag.Float64Var(&compressionRatio, "c", 8.0, "Compression ratio used for calculating disk space requirements.")
	flag.Float64Var(&compressionRatio, "compression-ratio", 8.0, "Compression ratio used for calculating disk space requirements.")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output.")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output.")
}

func main() {
	setupLogging()
	defer summaryLogger.Println("========== Done, program exiting. ==========")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	summaryLogger.Printf("========== SKY-FSBEXT version: %s by %s ==========\n", version, author)

	// Get the Windows version and build number
	osVersion, err := getWindowsVersion()
	if err != nil {
		summaryLogger.Printf("Operating system: %s\n", runtime.GOOS)
		summaryLogger.Printf("Failed to get Windows version: %v\n", err)
	} else {
		summaryLogger.Printf("Operating system: %s\n", osVersion)
	}

	if flag.Arg(0) == "--version" {
		fmt.Printf("SKY-FSBEXT version: %s by %s\n", version, author)
		return
	}

	inputDirSizeGB := float64(getSizeOfDir(inputDir)) / (1024 * 1024 * 1024)
	expectedSizeGB := inputDirSizeGB * compressionRatio
	expectedSizeBytes := uint64(expectedSizeGB * 1024 * 1024 * 1024)

	CheckDiskSpace(outputDir, expectedSizeBytes)

	log.Printf("Input directory: %s\n", inputDir)
	log.Printf("Output directory: %s\n", outputDir)

	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(inputDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create input directory: %v\n", err)
		}
		log.Println("Input directory not found - rebuilding")
	}

	bankFiles, err := filepath.Glob(filepath.Join(inputDir, "*.bank"))
	if err != nil {
		log.Fatalf("Failed to search for .bank files: %v\n", err)
	}

	if len(bankFiles) == 0 {
		log.Println("No sound banks found in input directory")
		flag.Usage()
		return
	}

	log.Printf("Found %d sound bank(s) in input directory\n", len(bankFiles))
	createDirectoryStructure(outputDir)

	if _, err := os.Stat(vgmstreamPath); os.IsNotExist(err) {
		log.Fatalf("vgmstream-cli executable not found at %s\n", vgmstreamPath)
	}

	if len(bankFiles) > 0 {
		extractedFiles := extractAndMoveFiles(bankFiles)

		if extractedFiles > 0 {
			log.Printf("Successfully extracted %d bank file(s)\n", extractedFiles)
		} else {
			log.Println("No sound banks were extracted")
		}

		removeEmptyDirectories(outputDir)
	}
}

func setupLogging() {
	log.SetFlags(log.LstdFlags)
	logFile, err := os.OpenFile("fsbext.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v\n", err)
	}

	// Create a custom logger for file logging
	fileLogger = log.New(logFile, "", log.LstdFlags)

	// Create a custom logger for console output
	consoleLogger = log.New(os.Stdout, "", 0)

	// Create a logger for summary messages that writes to both the console and the log file
	summaryLogger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)

	// Set the standard logger to write to the file
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func getSizeOfDir(directory string) int64 {
	var size int64
	err := filepath.Walk(directory, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to calculate directory size: %v\n", err)
	}
	return size
}

func createDirectoryStructure(outputDir string) {
	directories := []string{"Music", "SFX", "Other"}
	for _, dirName := range directories {
		dirPath := filepath.Join(outputDir, dirName)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			log.Printf("Failed to create directory %s: %v\n", dirName, err)
		} else {
			log.Printf("Created directory structure for %s.\n", dirName)
		}
	}
}

func removeEmptyDirectories(outputDir string) {
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			empty, err := isDirEmpty(path)
			if err != nil {
				return err
			}
			if empty {
				if err := os.RemoveAll(path); err != nil {
					log.Printf("Failed to remove directory: %s\n", path)
				} else {
					log.Printf("Removed empty directory: %s\n", path)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to remove empty directories: %v\n", err)
	}
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// getWindowsVersion retrieves the Windows version and build number
func getWindowsVersion() (string, error) {
	cmd := exec.Command("cmd", "/C", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func extractAndMoveFiles(bankFiles []string) int {
	extractedFiles := 0
	for i, bankFile := range bankFiles {
		if _, err := os.Stat(bankFile); os.IsNotExist(err) {
			fileLogger.Printf("Bank file does not exist: %s\n", bankFile)
			continue
		}

		if !isValidBankFile(bankFile) {
			fileLogger.Printf("Invalid bank file: %s\n", bankFile)
			continue
		}

		baseName := filepath.Base(bankFile)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		var bankDir string
		if strings.HasPrefix(baseName, "Music_") {
			bankDir = filepath.Join(outputDir, "Music", baseNameWithoutExt)
		} else if strings.HasPrefix(baseName, "SFX_") {
			bankDir = filepath.Join(outputDir, "SFX", baseNameWithoutExt)
		} else {
			bankDir = filepath.Join(outputDir, "Other", baseNameWithoutExt)
		}

		if err := os.MkdirAll(bankDir, os.ModePerm); err != nil {
			fileLogger.Printf("Failed to create or access directory %s: %v\n", bankDir, err)
			continue
		}

		outputPattern := filepath.Join(bankDir, "?n.wav")
		cmd := exec.Command(vgmstreamPath, "-v", "-o", outputPattern, bankFile)
		fmt.Printf("Processing file %d of %d: %s", i+1, len(bankFiles), bankFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(": FAIL\n")
			fileLogger.Printf("Failed to extract %s: %v\nCommand output: %s\n", bankFile, err, string(output))
			continue
		}

		extractedCount, err := countFilesInDir(bankDir)
		if err != nil {
			fmt.Printf(": FAIL\n")
			fileLogger.Printf("Error counting files in %s: %v\n", bankDir, err)
		} else if extractedCount == 0 {
			fmt.Printf(": FAIL\n")
			fileLogger.Printf("No files were extracted to %s\n", bankDir)
		} else {
			fmt.Printf(": OK\n")
			fileLogger.Printf("Successfully extracted %d files from %s to %s\n", extractedCount, bankFile, bankDir)
			extractedFiles += extractedCount
		}
	}
	return extractedFiles
}

func isValidBankFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fileLogger.Printf("Failed to open bank file %s: %v\n", filePath, err)
		return false
	}
	defer file.Close()

	// Read the first 4 bytes
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		fileLogger.Printf("Failed to read header of bank file %s: %v\n", filePath, err)
		return false
	}

	// Check if the header matches the expected format (you may need to adjust this)
	return string(header) == "RIFF" || string(header) == "FSB5"
}

// Helper function to count files in a directory
func countFilesInDir(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, file := range files {
		if !file.IsDir() {
			count++
		}
	}
	return count, nil
}
