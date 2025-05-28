# ImageMagick Client Package

This package provides a Go wrapper for the ImageMagick library using the `gopkg.in/gographics/imagick.v3` package.

## Features

- Image resizing with customizable dimensions
- Format conversion between different image formats
- Thread-safe operations with mutex locking
- Support for both file-based and in-memory operations

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
}
```

## Requirements

- Go 1.23.8 or higher
- ImageMagick library installed on the system

## Testing

Run the tests with:

```
go test -v ./pkg/mwclient
```

Note: Some tests require real image files and are skipped by default. See the test file for details on how to enable these tests.
