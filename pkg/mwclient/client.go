package mwclient

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// Common errors
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrProcessing   = errors.New("processing error")
)

// ImageMeta contains metadata about an image
type ImageMeta struct {
	FormatName      string
	ImageWidth      int32
	ImageHeight     int32
	ExifOrientation int16
	ContentLength   int64
}

// Client represents an ImageMagick client wrapper
type Client struct {
	mu sync.Mutex
}

// New creates a new ImageMagick client
func New() *Client {
	imagick.Initialize()
	return &Client{}
}

// Close releases resources used by the ImageMagick client
func (c *Client) Close() {
	imagick.Terminate()
}

// OpenImage opens an image from a file path and extracts metadata
func (c *Client) OpenImage(imagePath string) (ImageMeta, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var meta ImageMeta

	if imagePath == "" {
		return meta, fmt.Errorf("%w: image path is empty", ErrInvalidInput)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(imagePath)
	if err != nil {
		return meta, fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Extract metadata
	meta = ImageMeta{
		FormatName:      mw.GetImageFormat(),
		ImageWidth:      int32(mw.GetImageWidth()),
		ImageHeight:     int32(mw.GetImageHeight()),
		ExifOrientation: int16(mw.GetOrientation()),
	}

	if cl, err := mw.GetImageLength(); err == nil {
		meta.ContentLength = int64(cl)
	}

	// Auto-orient the image based on EXIF data
	err = mw.AutoOrientImage()
	if err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	return meta, nil
}

// GetImageMetadata extracts metadata from an image file
func (c *Client) GetImageMetadata(imagePath string) (ImageMeta, error) {
	// This is now just an alias for OpenImage for backward compatibility
	return c.OpenImage(imagePath)
}

// ResizeImage resizes an image from a reader to the specified dimensions
// and writes the result to the provided writer
func (c *Client) ResizeImage(r io.Reader, w io.Writer, width, height uint, format string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if r == nil || w == nil {
		return fmt.Errorf("%w: reader or writer is nil", ErrInvalidInput)
	}

	if width == 0 || height == 0 {
		return fmt.Errorf("%w: invalid dimensions", ErrInvalidInput)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read image data
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	if err := mw.ReadImageBlob(data); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	if err := mw.AutoOrientImage(); err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	// Resize the image using the Sinc filter (as in the original implementation)
	if err := mw.ResizeImage(width, height, imagick.FILTER_SINC); err != nil {
		return fmt.Errorf("%w: failed to resize image: %v", ErrProcessing, err)
	}

	// Set compression quality to 95 (high quality)
	if err := mw.SetImageCompressionQuality(95); err != nil {
		return fmt.Errorf("%w: failed to set compression quality: %v", ErrProcessing, err)
	}

	// Set the output format if specified
	if format != "" {
		if err := mw.SetImageFormat(format); err != nil {
			return fmt.Errorf("%w: failed to set image format: %v", ErrProcessing, err)
		}
	}

	// Get the image blob
	blob, err := mw.GetImageBlob()
	if err != nil {
		return fmt.Errorf("%w: failed to get image blob: %v", ErrProcessing, err)
	}
	if len(blob) == 0 {
		return fmt.Errorf("%w: empty result image", ErrProcessing)
	}

	// Write the result
	if _, err := w.Write(blob); err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// ResizeImageFile resizes an image from a file path to the specified dimensions
// and writes the result to the output file path
func (c *Client) ResizeImageFile(inputPath, outputPath string, width, height uint, format string) error {
	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("%w: input or output path is empty", ErrInvalidInput)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image
	if err := mw.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	if err := mw.AutoOrientImage(); err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	// Resize the image using the Sinc filter
	if err := mw.ResizeImage(width, height, imagick.FILTER_SINC); err != nil {
		return fmt.Errorf("%w: failed to resize image: %v", ErrProcessing, err)
	}

	// Set compression quality to 95 (high quality)
	if err := mw.SetImageCompressionQuality(95); err != nil {
		return fmt.Errorf("%w: failed to set compression quality: %v", ErrProcessing, err)
	}

	// Set the output format if specified
	if format != "" {
		if err := mw.SetImageFormat(format); err != nil {
			return fmt.Errorf("%w: failed to set image format: %v", ErrProcessing, err)
		}
	}

	// Write the image directly to file
	if err := mw.WriteImage(outputPath); err != nil {
		return fmt.Errorf("%w: failed to write image: %v", ErrProcessing, err)
	}

	return nil
}

// ConvertFormat converts an image from one format to another
func (c *Client) ConvertFormat(r io.Reader, w io.Writer, format string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if r == nil || w == nil {
		return fmt.Errorf("%w: reader or writer is nil", ErrInvalidInput)
	}

	if format == "" {
		return fmt.Errorf("%w: format is empty", ErrInvalidInput)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read image data
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	if err := mw.ReadImageBlob(data); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	if err := mw.AutoOrientImage(); err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	// Set compression quality to 95 (high quality)
	if err := mw.SetImageCompressionQuality(95); err != nil {
		return fmt.Errorf("%w: failed to set compression quality: %v", ErrProcessing, err)
	}

	// Set the output format
	if err := mw.SetImageFormat(format); err != nil {
		return fmt.Errorf("%w: failed to set image format: %v", ErrProcessing, err)
	}

	// Get the image blob
	blob, err := mw.GetImageBlob()
	if err != nil {
		return fmt.Errorf("%w: failed to get image blob: %v", ErrProcessing, err)
	}
	if len(blob) == 0 {
		return fmt.Errorf("%w: empty result image", ErrProcessing)
	}

	// Write the result
	if _, err := w.Write(blob); err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// ResizeByHeight resizes an image to a specific height while maintaining aspect ratio
func (c *Client) ResizeByHeight(inputPath, outputPath string, targetHeight int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("%w: input or output path is empty", ErrInvalidInput)
	}

	if targetHeight <= 0 {
		return fmt.Errorf("%w: target height must be positive", ErrInvalidInput)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image
	if err := mw.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	if err := mw.AutoOrientImage(); err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	// Get image dimensions
	imageWidth := int32(mw.GetImageWidth())
	imageHeight := int32(mw.GetImageHeight())

	slog.Info("ResizeByHeight with Sinc", "Out", outputPath, "Height", targetHeight)
	
	// Calculate the target width, keeping aspect ratio
	targetWidth := uint(imageWidth * int32(targetHeight) / imageHeight)

	// Resize the image using the Sinc filter
	if err := mw.ResizeImage(targetWidth, uint(targetHeight), imagick.FILTER_SINC); err != nil {
		return fmt.Errorf("%w: failed to resize image: %v", ErrProcessing, err)
	}

	// Set compression quality to 95 (high quality)
	if err := mw.SetImageCompressionQuality(95); err != nil {
		return fmt.Errorf("%w: failed to set compression quality: %v", ErrProcessing, err)
	}

	// Write the image directly to file
	if err := mw.WriteImage(outputPath); err != nil {
		return fmt.Errorf("%w: failed to write image: %v", ErrProcessing, err)
	}

	return nil
}

// ResizeByWidth resizes an image to a specific width while maintaining aspect ratio
func (c *Client) ResizeByWidth(inputPath, outputPath string, targetWidth int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("%w: input or output path is empty", ErrInvalidInput)
	}

	if targetWidth <= 0 {
		return fmt.Errorf("%w: target width must be positive", ErrInvalidInput)
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: input file does not exist: %v", ErrInvalidInput, err)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image
	if err := mw.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	if err := mw.AutoOrientImage(); err != nil {
		slog.Error("Auto-orientation failed", "error", err)
		// Continue despite error
	}

	// Get image dimensions
	imageWidth := int32(mw.GetImageWidth())
	imageHeight := int32(mw.GetImageHeight())

	slog.Info("ResizeByWidth with Sinc", "Out", outputPath, "Width", targetWidth)
	
	// Calculate the target height, keeping aspect ratio
	targetHeight := uint(imageHeight * int32(targetWidth) / imageWidth)

	// Resize the image using the Sinc filter
	if err := mw.ResizeImage(uint(targetWidth), targetHeight, imagick.FILTER_SINC); err != nil {
		return fmt.Errorf("%w: failed to resize image: %v", ErrProcessing, err)
	}

	// Set compression quality to 95 (high quality)
	if err := mw.SetImageCompressionQuality(95); err != nil {
		return fmt.Errorf("%w: failed to set compression quality: %v", ErrProcessing, err)
	}

	// Write the image directly to file
	if err := mw.WriteImage(outputPath); err != nil {
		return fmt.Errorf("%w: failed to write image: %v", ErrProcessing, err)
	}

	return nil
}
