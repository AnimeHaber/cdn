package main

import (
	"cdn-service/config"
	"cdn-service/handlers"
	"log"
	"strconv"
	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // Default 10MB limit
	})

	// Middleware
	app.Use(logger.New())

	// Security Headers (Helmet)
	app.Use(helmet.New())

	// Rate Limiting (60 requests per minute)
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-API-KEY", c.IP()) // Limit by API Key if present, else IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please slow down.",
			})
		},
	}))

	// Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("CDN Service Operational")
	})

	// Routing Group for Uploads
	// Secured by API Key
	api := app.Group("/api", handlers.CheckAPIKey(cfg))
	api.Post("/upload/avatar", handlers.Upload(cfg, "avatar"))
	api.Post("/upload/image", handlers.Upload(cfg, "image"))

	// Static File Serving (CDN)
	// Cache Configuration
	cacheSeconds, _ := strconv.Atoi(cfg.CacheDuration)
	if cacheSeconds <= 0 {
		cacheSeconds = 31536000 // 1 year default
	}

	staticConfig := fiber.Static{
		Compress:      true,
		Browse:        false,
		CacheDuration: time.Duration(cacheSeconds) * time.Second,
		MaxAge:        cacheSeconds,
	}

	// Serve avatar and image directories
	// URLs will be like: http://cdn-host/avatar/5-5-5-5.jpg
	app.Static("/avatar", cfg.StorageAvatar, staticConfig)
	app.Static("/image", cfg.StorageImage, staticConfig)

	// Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
