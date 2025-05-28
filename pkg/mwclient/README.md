# ImageMagick Client Package

This package provides a Go wrapper for the ImageMagick library using the `gopkg.in/gographics/imagick.v3` package.

## Features

- Image resizing with customizable dimensions
- Format conversion between different image formats
- Thread-safe operations with mutex locking
- Support for both file-based and in-memory operations
- Image metadata extraction and access
- Automatic image orientation based on EXIF data
- High-quality image compression (95% quality)
- Aspect ratio-preserving resize operations
- PDF to image conversion with montage support

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/torpago/simple-media-proc/pkg/mwclient"
)

func main() {
	// Create a new client
	client := mwclient.New()
	defer client.Close()

	// Resize an image file
	err := client.ResizeImageFile("input.jpg", "output.png", 800, 600, "png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Image resized successfully!")
	
	// Resize an image while preserving aspect ratio
	err = client.ResizeByHeight("input.jpg", "output_height.png", 600)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Get image metadata
	meta, err := client.OpenImage("input.jpg")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Image format: %s, dimensions: %dx%d\n", 
		meta.FormatName, meta.ImageWidth, meta.ImageHeight)
	
	// Create a montage of the first 2 pages of a PDF
	err = client.ConvertPdfToImages("input.pdf", "output.png", 2, 300, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Or extract each page as a separate file
	err = client.ConvertPdfToImages("input.pdf", "output.png", 2, 300, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Requirements

- Go 1.23.8 or higher
- ImageMagick library installed on the system

## Testing

Run the tests with:

```
make test
```

Note: Some tests require real image files and are skipped by default. See the test file for details on how to enable these tests.
