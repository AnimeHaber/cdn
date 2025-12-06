package handlers

import (
	"cdn-service/config"
	"cdn-service/utils"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CheckAPIKey matches the header against config
func CheckAPIKey(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-KEY")
		if key == "" || key != cfg.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized access",
			})
		}
		return c.Next()
	}
}

// Upload handles the logic for saving files with generated names
func Upload(cfg *config.Config, category string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse file
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "File is required (key: 'file')",
			})
		}

		// 1. Validate Extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		validExts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true,
			".gif": true, ".webp": true, ".svg": true,
		}

		if !validExts[ext] {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "File extension not allowed",
			})
		}

		// 2. Validate Content (Magic Bytes)
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to open file for validation",
			})
		}

		buffer := make([]byte, 512)
		_, err = fileContent.Read(buffer)
		fileContent.Close()

		if err != nil && err.Error() != "EOF" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read file content",
			})
		}

		contentType := http.DetectContentType(buffer)
		validMimes := map[string]bool{
			"image/jpeg": true, "image/png": true,
			"image/gif": true, "image/webp": true, "image/svg+xml": true,
			"text/plain; charset=utf-8": true,
		}

		isSVG := ext == ".svg" && (strings.Contains(contentType, "text/xml") || strings.Contains(contentType, "text/plain"))

		if !validMimes[contentType] && !isSVG {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "Real file content verification failed. Detected: " + contentType,
			})
		}

		// Generate Random Name: 5-5-5-5 format
		newFileName := utils.GenerateCDNFileName() + ext

		// Determine Storage Path
		var targetDir string
		if category == "avatar" {
			targetDir = cfg.StorageAvatar
		} else {
			targetDir = cfg.StorageImage
		}

		// Save File
		destination := filepath.Join(targetDir, newFileName)
		if err := c.SaveFile(file, destination); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not save file to disk",
			})
		}

		// Return success with name and category type
		return c.JSON(fiber.Map{
			"success": true,
			"name":    newFileName,
			"type":    category,
		})
	}
}
