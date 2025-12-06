package handlers

import (
	"cdn-service/config"
	"cdn-service/utils"
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

		// Validate extension (Basic Image Check)
		ext := strings.ToLower(filepath.Ext(file.Filename))
		validExts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, 
			".gif": true, ".webp": true, ".svg": true,
		}
		
		if !validExts[ext] {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "Only image files are allowed",
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

		// Return success with just the filename as the user might want to construct URL themselves
		// or return relative path. User said "bu isim dönecek" (this name will return).
		return c.JSON(fiber.Map{
			"success": true,
			"name":    newFileName,
			"path":    "/" + category + "/" + newFileName,
		})
	}
}
