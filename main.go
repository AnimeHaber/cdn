package main

import (
	"cdn-service/config"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize Config
	cfg := config.LoadConfig()

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // Default 10MB limit (configurable later)
	})

	// Middleware
	app.Use(logger.New())

	// Routing Group for Uploads
	api := app.Group("/api", handlers.CheckAPIKey(cfg))
	
	// Upload Endpoints
	api.Post("/upload/avatar", handlers.Upload(cfg, "avatar"))
	api.Post("/upload/image", handlers.Upload(cfg, "image"))

	// Static File Serving (The CDN part)
	// We allow public access to the data folders directly
	// Cache Control: aggressive caching as requested
	cacheTime := 3600 * 24 * 365 // Default 1 year roughly if parsing fails, but fiber handles Duration
	
	// Convert env string to duration (simplified, using default fiber int which is seconds usually or time.Duration)
	// Fiber Static CacheDuration is time.Duration.
	// We'll trust the integer in env is seconds, but Fiber Static expects duration.
	// Actually Fiber Static CacheDuration matches the header "Cache-Control: public, max-age=XXX"
	
	app.Static("/data/avatar", cfg.StorageAvatar, fiber.Static{
		Compress:      true,
		Browse:        false,
		CacheDuration: -1, // We will set custom header manually if needed, or rely on MaxAge 
		// Fiber v2 Static has MaxAge int (seconds) ? No, CacheDuration time.Duration
	})
	
	// Let's use a custom handler for static to ensure we get exactly the headers we want for a CDN
	// Or use Fiber's Static correctly:
	// "CacheDuration" sets the expiration.
	
	import (
		"cdn-service/handlers"
		"time"
		"strconv"
	)
	
	cacheSeconds, _ := strconv.Atoi(cfg.CacheDuration)
	if cacheSeconds <= 0 {
		cacheSeconds = 31536000 // 1 year
	}
	
	staticConfig := fiber.Static{
		Compress:      true,
		CacheDuration: time.Duration(cacheSeconds) * time.Second,
		MaxAge:        cacheSeconds, // Cache-Control: max-age=...
	}

	app.Static("/avatar", cfg.StorageAvatar, staticConfig)
	app.Static("/image", cfg.StorageImage, staticConfig)

	// Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
