package mwclient

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Fatal("New() returned nil")
	}
	defer client.Close()
}

func TestResizeImage(t *testing.T) {
	client := New()
	defer client.Close()

	tests := []struct {
		name      string
		reader    io.Reader
		width     uint
		height    uint
		format    string
		expectErr bool
	}{
		{
			name:      "nil reader",
			reader:    nil,
			width:     100,
			height:    100,
			format:    "png",
			expectErr: true,
		},
		{
			name:      "zero width",
			reader:    bytes.NewReader([]byte("test")),
			width:     0,
			height:    100,
			format:    "png",
			expectErr: true,
		},
		{
			name:      "zero height",
			reader:    bytes.NewReader([]byte("test")),
			width:     100,
			height:    0,
			format:    "png",
			expectErr: true,
		},
		{
			name:      "invalid image data",
			reader:    bytes.NewReader([]byte("not an image")),
			width:     100,
			height:    100,
			format:    "png",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := client.ResizeImage(tt.reader, &buf, tt.width, tt.height, tt.format)
			
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvertFormat(t *testing.T) {
	client := New()
	defer client.Close()

	tests := []struct {
		name      string
		reader    io.Reader
		format    string
		expectErr bool
	}{
		{
			name:      "nil reader",
			reader:    nil,
			format:    "png",
			expectErr: true,
		},
		{
			name:      "empty format",
			reader:    bytes.NewReader([]byte("test")),
			format:    "",
			expectErr: true,
		},
		{
			name:      "invalid image data",
			reader:    bytes.NewReader([]byte("not an image")),
			format:    "png",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := client.ConvertFormat(tt.reader, &buf, tt.format)
			
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestWithRealImage is a helper function to test with a real image file
// This test is skipped by default as it requires a real image file
func TestWithRealImage(t *testing.T) {
	t.Skip("Skipping test that requires a real image file")
	
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "mwclient-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Path to a test image (you would need to provide this)
	testImagePath := "testdata/test.jpg"
	
	// Output path
	outputPath := filepath.Join(tempDir, "output.png")
	
	client := New()
	defer client.Close()
	
	// Test ResizeImageFile
	err = client.ResizeImageFile(testImagePath, outputPath, 200, 200, "png")
	if err != nil {
		t.Errorf("ResizeImageFile failed: %v", err)
	}
	
	// Check if the output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}
}

// TestResizeImageFile tests the error cases for ResizeImageFile
func TestResizeImageFile(t *testing.T) {
	client := New()
	defer client.Close()
	
	// Test with empty paths
	err := client.ResizeImageFile("", "output.png", 100, 100, "png")
	if err == nil {
		t.Error("Expected error with empty input path, got nil")
	}
	
	err = client.ResizeImageFile("input.png", "", 100, 100, "png")
	if err == nil {
		t.Error("Expected error with empty output path, got nil")
	}
	
	// Test with non-existent input file
	err = client.ResizeImageFile("nonexistent.png", "output.png", 100, 100, "png")
	if err == nil {
		t.Error("Expected error with non-existent input file, got nil")
	}
}
