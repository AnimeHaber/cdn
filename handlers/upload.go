package handlers

import (
	"cdn-service/config"
	"path/filepath"

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
