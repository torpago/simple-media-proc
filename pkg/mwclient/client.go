package mwclient

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// Common errors
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrProcessing   = errors.New("processing error")
)

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

	// Resize the image
	if err := mw.ResizeImage(width, height, imagick.FILTER_LANCZOS); err != nil {
		return fmt.Errorf("%w: failed to resize image: %v", ErrProcessing, err)
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

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	return c.ResizeImage(inputFile, outputFile, width, height, format)
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
