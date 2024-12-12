package cache

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/nfnt/resize"
)

const (
	maxImageDimension = 800
	jpegQuality      = 85
)

func (m *ImageCacheManager) loadAndOptimizeImage(path string) (fyne.Resource, error) {
	// Read the image file
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
	}
	defer imgFile.Close()

	// Decode the image
	img, format, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	var newWidth, newHeight uint
	if width > height {
		if width > maxImageDimension {
			newWidth = maxImageDimension
			newHeight = uint(float64(height) * (float64(maxImageDimension) / float64(width)))
		} else {
			newWidth = uint(width)
			newHeight = uint(height)
		}
	} else {
		if height > maxImageDimension {
			newHeight = maxImageDimension
			newWidth = uint(float64(width) * (float64(maxImageDimension) / float64(height)))
		} else {
			newWidth = uint(width)
			newHeight = uint(height)
		}
	}

	// Resize the image using Lanczos resampling
	resized := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// Create a buffer to store the optimized image
	var buf bytes.Buffer

	// Encode based on original format
	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %v", err)
		}
	case "png":
		if err := png.Encode(&buf, resized); err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	// Create a static resource
	return fyne.NewStaticResource(filepath.Base(path), buf.Bytes()), nil
}
