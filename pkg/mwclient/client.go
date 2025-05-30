package mwclient

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	slog.Info("ReadImage", "In", inputPath)
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
		slog.Info("SetImageFormat", "Format", format)
		if err := mw.SetImageFormat(format); err != nil {
			return fmt.Errorf("%w: failed to set image format: %v", ErrProcessing, err)
		}
	}

	// Write the image directly to file
	slog.Info("WriteImage", "Out", outputPath)
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
	slog.Info("ReadImage", "In", inputPath)
	if err := mw.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	slog.Info("AutoOrientImage")
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
	slog.Info("WriteImage", "Out", outputPath)
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
	slog.Info("ReadImage", "In", inputPath)
	if err := mw.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read image: %v", ErrProcessing, err)
	}

	// Auto-orient the image based on EXIF data
	slog.Info("AutoOrientImage")
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
	slog.Info("WriteImage", "Out", outputPath)
	if err := mw.WriteImage(outputPath); err != nil {
		return fmt.Errorf("%w: failed to write image: %v", ErrProcessing, err)
	}

	return nil
}

// ConvertPdfToImages converts a PDF file to one or more images
// If createMontage is true, it will combine the images into a single montage image
// maxPages limits the number of pages to process (0 means all pages)
// targetHeight specifies the height for the output images
func (c *Client) ConvertPdfToImages(inputPath, outputPath string, maxPages int, targetHeight int, createMontage bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("%w: input or output path is empty", ErrInvalidInput)
	}

	if targetHeight <= 0 {
		return fmt.Errorf("%w: target height must be positive", ErrInvalidInput)
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: input file does not exist: %v", ErrInvalidInput, err)
	}

	// Read the PDF
	pdfWand := imagick.NewMagickWand()
	defer pdfWand.Destroy()

	// bump PDF raster density to 300 DPI for sharper text/lines:
	if err := pdfWand.SetResolution(300, 300); err != nil {
		return fmt.Errorf("%w: could not set resolution: %v", ErrProcessing, err)
	}

	if err := pdfWand.ReadImage(inputPath); err != nil {
		return fmt.Errorf("%w: failed to read PDF: %v", ErrProcessing, err)
	}

	// Get the number of pages
	numPages := pdfWand.GetNumberImages()
	slog.Info("ConvertPdf", "Out", outputPath, "Page Height", targetHeight, "Total Pages", numPages)

	// Limit the number of pages if maxPages is specified
	if maxPages > 0 && int(numPages) > maxPages {
		numPages = uint(maxPages)
	}

	// Create a new wand for the output images
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Add each page to the output wand
	for i := 0; i < int(numPages); i++ {
		slog.Info("Processing page", "Index", i)
		pdfWand.SetIteratorIndex(i)
		pageImg := pdfWand.GetImage()

		// Add the page image to the output wand
		err := mw.AddImage(pageImg)
		if err != nil {
			slog.Error("Failed to add page image", "error", err, "page", i)
			continue
		}

		// If not creating a montage, save each page as a separate file
		if !createMontage {
			// Get the current image from the wand
			mw.SetIteratorIndex(i)
			currentImg := mw.GetImage()

			// flatten transparency over white
			white := imagick.NewPixelWand()
			defer white.Destroy()
			white.SetColor("white")
			currentImg.SetImageBackgroundColor(white)
			flat := currentImg.MergeImageLayers(imagick.IMAGE_LAYER_FLATTEN) // new flat wand
			currentImg.Destroy()                                             // drop the raw one early
			currentImg = flat                                                // now work with flat version
			defer currentImg.Destroy()                                       // clean up

			// Auto-orient the image based on EXIF data
			err := currentImg.AutoOrientImage()
			if err != nil {
				slog.Error("Auto-orientation failed", "error", err)
				// Continue despite error
			}

			// Resize to the target height
			imageWidth := int32(currentImg.GetImageWidth())
			imageHeight := int32(currentImg.GetImageHeight())
			targetWidth := uint(imageWidth * int32(targetHeight) / imageHeight)

			if err := currentImg.ResizeImage(targetWidth, uint(targetHeight), imagick.FILTER_SINC); err != nil {
				slog.Error("Failed to resize page image", "error", err, "page", i)
				continue
			}

			// Set compression quality
			if err := currentImg.SetImageCompressionQuality(95); err != nil {
				slog.Error("Failed to set compression quality", "error", err, "page", i)
			}

			// Generate the output filename for this page
			pageOutputPath := outputPath
			if numPages > 1 {
				ext := filepath.Ext(outputPath)
				base := strings.TrimSuffix(outputPath, ext)
				pageOutputPath = fmt.Sprintf("%s_page%d%s", base, i+1, ext)
				slog.Info("Processing page", "Index", i, "Path", pageOutputPath)
			}

			// Write the page image to file
			if err := currentImg.WriteImage(pageOutputPath); err != nil {
				slog.Error("Failed to write page image", "error", err, "page", i, "path", pageOutputPath)
			}

			slog.Info("Processed page", "Index", i)
		}
	}

	// If creating a montage, combine all pages into one image
	if createMontage {
		// Create a drawing wand for the montage
		dw := imagick.NewDrawingWand()
		defer dw.Destroy()

		// Set up montage parameters
		tileGeo := "1x"                              // Stack vertically
		thumbGeo := fmt.Sprintf("x%d", targetHeight) // Target height
		mode := imagick.MONTAGE_MODE_CONCATENATE
		frame := "+0+0" // No frame

		// Create the montage
		montageWand := mw.MontageImage(dw, tileGeo, thumbGeo, mode, frame)
		defer montageWand.Destroy()

		// Set compression quality
		if err := montageWand.SetImageCompressionQuality(95); err != nil {
			slog.Error("Failed to set montage compression quality", "error", err)
		}

		// Write the montage to file
		if err := montageWand.WriteImage(outputPath); err != nil {
			return fmt.Errorf("%w: failed to write montage image: %v", ErrProcessing, err)
		}
	}

	return nil
}
