package prism

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidationResult contains detailed validation information
type ValidationResult struct {
	Valid        bool
	Errors       []string
	Warnings     []string
	Capabilities []string
}

// Validate performs comprehensive prism binary validation
func Validate(binaryPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:        true,
		Errors:       []string{},
		Warnings:     []string{},
		Capabilities: []string{},
	}

	// Check 1: File exists
	info, err := os.Stat(binaryPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("File not found: %v", err))
		return result, nil
	}

	// Check 2: Executable permissions
	if !isExecutable(binaryPath) {
		result.Valid = false
		result.Errors = append(result.Errors, "File is not executable")
		return result, nil
	}

	// Check 3: File size warning
	const maxRecommendedSize = 100 * 1024 * 1024 // 100MB
	if info.Size() > maxRecommendedSize {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Large binary size: %d MB", info.Size()/(1024*1024)))
	}

	// Check 4: Detect if it's a script
	if isScript(binaryPath) {
		result.Warnings = append(result.Warnings, "Binary appears to be a script (shebang detected)")
	}

	// Check 5: Try to detect ELF binary type
	capabilities, err := detectCapabilities(binaryPath)
	if err == nil {
		result.Capabilities = capabilities
	}

	// Check 6: Try running with --version
	versionOutput, err := checkVersion(binaryPath)
	if err == nil && versionOutput != "" {
		result.Capabilities = append(result.Capabilities, "version-flag")
	}

	return result, nil
}

// isScript checks if a file starts with a shebang
func isScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 2)
	n, err := f.Read(buf)
	if err != nil || n < 2 {
		return false
	}

	return buf[0] == '#' && buf[1] == '!'
}

// detectCapabilities attempts to detect binary capabilities
func detectCapabilities(path string) ([]string, error) {
	capabilities := []string{}

	f, err := elf.Open(path)
	if err != nil {
		return capabilities, err
	}
	defer f.Close()

	// Basic ELF info
	capabilities = append(capabilities, fmt.Sprintf("arch:%s", f.Machine.String()))
	capabilities = append(capabilities, fmt.Sprintf("type:%s", f.Type.String()))

	// Check if dynamically linked
	if f.Type == elf.ET_DYN {
		capabilities = append(capabilities, "dynamic")
	} else if f.Type == elf.ET_EXEC {
		capabilities = append(capabilities, "static")
	}

	return capabilities, nil
}

// checkVersion tries to run the binary with --version
func checkVersion(path string) (string, error) {
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateManifest validates a prism manifest and its referenced binary
func ValidateManifest(manifestPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:        true,
		Errors:       []string{},
		Warnings:     []string{},
		Capabilities: []string{},
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load manifest: %v", err))
		return result, nil
	}

	if err := manifest.Validate(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid manifest: %v", err))
		return result, nil
	}

	// Validate referenced binary
	binaryPath := manifest.Prism.Binary
	if !strings.Contains(binaryPath, string(os.PathSeparator)) {
		manifestDir := filepath.Dir(manifestPath)
		binaryPath = filepath.Join(manifestDir, binaryPath)
	}

	binaryResult, err := Validate(binaryPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Binary validation error: %v", err))
		result.Valid = false
		return result, nil
	}

	// Merge binary validation results
	result.Errors = append(result.Errors, binaryResult.Errors...)
	result.Warnings = append(result.Warnings, binaryResult.Warnings...)
	result.Capabilities = append(result.Capabilities, binaryResult.Capabilities...)
	result.Valid = result.Valid && binaryResult.Valid

	return result, nil
}
