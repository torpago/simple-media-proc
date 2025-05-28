package mwclient

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func TestNew(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	client := New()
	if client == nil {
		t.Fatal("New() returned nil")
	}
	defer client.Close()
}

// Helper function to check if ImageMagick is available
func isImageMagickAvailable() bool {
	defer func() {
		// Recover from any panic that might occur when checking ImageMagick
		if r := recover(); r != nil {
			// ImageMagick not available
		}
	}()
	
	// Try to initialize ImageMagick
	imagick.Initialize()
	defer imagick.Terminate()
	
	// If we get here, ImageMagick is available
	return true
}

func TestResizeImage(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

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
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

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
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	// Skip if test image doesn't exist
	testImagePath := "/Users/bd/Workspace/Torpago/simple-media-proc/test/data/IMG_9908.jpeg"
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skip("Test image not found, skipping test")
	}
	
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "mwclient-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
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
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

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

// TestOpenImage tests the OpenImage method
func TestOpenImage(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	client := New()
	defer client.Close()
	
	// Test with empty path
	_, err := client.OpenImage("")
	if err == nil {
		t.Error("Expected error with empty path, got nil")
	}
	
	// Test with non-existent file
	_, err = client.OpenImage("nonexistent.png")
	if err == nil {
		t.Error("Expected error with non-existent file, got nil")
	}
}

// TestGetImageMetadata tests the GetImageMetadata method
func TestGetImageMetadata(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	client := New()
	defer client.Close()
	
	// Test with empty path
	_, err := client.GetImageMetadata("")
	if err == nil {
		t.Error("Expected error with empty path, got nil")
	}
	
	// Test with non-existent file
	_, err = client.GetImageMetadata("nonexistent.png")
	if err == nil {
		t.Error("Expected error with non-existent file, got nil")
	}
}

// TestResizeByHeight tests the ResizeByHeight method
func TestResizeByHeight(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	client := New()
	defer client.Close()
	
	// Test with empty paths
	err := client.ResizeByHeight("", "output.png", 100)
	if err == nil {
		t.Error("Expected error with empty input path, got nil")
	}
	
	err = client.ResizeByHeight("input.png", "", 100)
	if err == nil {
		t.Error("Expected error with empty output path, got nil")
	}
	
	// Test with invalid height
	err = client.ResizeByHeight("input.png", "output.png", 0)
	if err == nil {
		t.Error("Expected error with zero height, got nil")
	}
	
	err = client.ResizeByHeight("input.png", "output.png", -1)
	if err == nil {
		t.Error("Expected error with negative height, got nil")
	}
	
	// Test with non-existent input file
	err = client.ResizeByHeight("nonexistent.png", "output.png", 100)
	if err == nil {
		t.Error("Expected error with non-existent input file, got nil")
	}
}

// TestResizeByWidth tests the ResizeByWidth method
func TestResizeByWidth(t *testing.T) {
	// Skip test if ImageMagick is not properly configured
	if !isImageMagickAvailable() {
		t.Skip("ImageMagick not available, skipping test")
	}

	client := New()
	defer client.Close()
	
	// Test with empty paths
	err := client.ResizeByWidth("", "output.png", 100)
	if err == nil {
		t.Error("Expected error with empty input path, got nil")
	}
	
	err = client.ResizeByWidth("input.png", "", 100)
	if err == nil {
		t.Error("Expected error with empty output path, got nil")
	}
	
	// Test with invalid width
	err = client.ResizeByWidth("input.png", "output.png", 0)
	if err == nil {
		t.Error("Expected error with zero width, got nil")
	}
	
	err = client.ResizeByWidth("input.png", "output.png", -1)
	if err == nil {
		t.Error("Expected error with negative width, got nil")
	}
	
	// Test with non-existent input file
	err = client.ResizeByWidth("nonexistent.png", "output.png", 100)
	if err == nil {
		t.Error("Expected error with non-existent input file, got nil")
	}
}
