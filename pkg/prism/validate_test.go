package prism

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create executable file
	execPath := filepath.Join(tmpDir, "test-binary")
	content := []byte("#!/bin/sh\necho test\n")
	if err := os.WriteFile(execPath, content, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	result, err := Validate(execPath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result, got invalid with errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning about script shebang")
	}
}

func TestValidateNonExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-executable file
	filePath := filepath.Join(tmpDir, "test-file")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := Validate(filePath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for non-executable file")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error about non-executable file")
	}
}

func TestValidateNonExistent(t *testing.T) {
	result, err := Validate("/nonexistent/path/binary")
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for nonexistent file")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error about nonexistent file")
	}
}

func TestIsScript(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "shell script",
			content:  []byte("#!/bin/sh\necho test"),
			expected: true,
		},
		{
			name:     "python script",
			content:  []byte("#!/usr/bin/env python\nprint('test')"),
			expected: true,
		},
		{
			name:     "binary data",
			content:  []byte("\x7fELF\x02\x01\x01\x00"),
			expected: false,
		},
		{
			name:     "text file",
			content:  []byte("Some text content"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name)
			if err := os.WriteFile(path, tt.content, 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := isScript(path)
			if result != tt.expected {
				t.Errorf("Expected isScript=%v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLargeBinaryWarning(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a large file (over 100MB)
	largePath := filepath.Join(tmpDir, "large-binary")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	defer f.Close()

	// Write 101MB of data
	chunk := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 101; i++ {
		if _, err := f.Write(chunk); err != nil {
			t.Fatalf("Failed to write chunk: %v", err)
		}
	}

	if err := os.Chmod(largePath, 0755); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	result, err := Validate(largePath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	foundWarning := false
	for _, warning := range result.Warnings {
		if len(warning) > 0 && warning[:5] == "Large" {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Error("Expected warning about large binary size")
	}
}
