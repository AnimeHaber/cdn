package handlers

import (
	"cdn-service/config"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
)

// ServeOptimizedImage handles dynamic image resonse with caching
func ServeOptimizedImage(cfg *config.Config) fiber.Handler {
	// Ensure cache directory exists
	cacheDir := filepath.Join("data", "cache")
	os.MkdirAll(cacheDir, 0755)

	return func(c *fiber.Ctx) error {
		category := c.Params("category") // avatar or image
		filename := c.Params("filename")

		// 1. Basic Validation
		if category != "avatar" && category != "image" {
			return c.SendStatus(fiber.StatusNotFound)
		}

		// Security: Prevent directory traversal
		if filepath.Base(filename) != filename {
			return c.SendStatus(fiber.StatusForbidden)
		}

		// determine Source directory
		var srcDir string
		if category == "avatar" {
			srcDir = cfg.StorageAvatar
		} else {
			srcDir = cfg.StorageImage
		}

		srcPath := filepath.Join(srcDir, filename)

		// Check if source exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			return c.SendStatus(fiber.StatusNotFound)
		}

		// 2. Parse Query Parameters
		widthVal := c.Query("w")
		heightVal := c.Query("h")
		qualityVal := c.Query("q")

		// If no optimization requested, serve original file directly
		if widthVal == "" && heightVal == "" && qualityVal == "" {
			return c.SendFile(srcPath)
		}

		// 3. Optimization Logic
		width, _ := strconv.Atoi(widthVal)
		height, _ := strconv.Atoi(heightVal)
		quality, _ := strconv.Atoi(qualityVal)
		if quality == 0 {
			quality = 80
		} // Default quality
		if quality > 100 {
			quality = 100
		}

		// Generate a unique cache key/filename
		// format: filename_wX_hY_qZ.ext
		ext := filepath.Ext(filename)
		baseName := filename[:len(filename)-len(ext)]
		cacheFileName := fmt.Sprintf("%s_w%d_h%d_q%d%s", baseName, width, height, quality, ext)
		cachePath := filepath.Join(cacheDir, cacheFileName)

		// 4. Check Cache
		// If cached file exists, serve it immediately
		if _, err := os.Stat(cachePath); err == nil {
			c.Set("X-Cache", "HIT")
			return c.SendFile(cachePath)
		}

		// 5. Process Image (Cache Miss)
		// Load original
		srcParams, err := imaging.Open(srcPath)
		if err != nil {
			// If file is not an image (e.g. svg), serve original
			return c.SendFile(srcPath)
		}

		// Resize if requested
		var processedImg image.Image = srcParams
		if width > 0 || height > 0 {
			// specific resize using Lanczos for best quality
			// 0 means preserve aspect ratio
			processedImg = imaging.Resize(processedImg, width, height, imaging.Lanczos)
		}

		// Save to Cache
		err = imaging.Save(processedImg, cachePath, imaging.JPEGQuality(quality))
		if err != nil {
			// If save fails, just serve original as fallback
			return c.SendFile(srcPath)
		}

		c.Set("X-Cache", "MISS")
		return c.SendFile(cachePath)
	}
}
