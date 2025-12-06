package main

import (
	"cdn-service/config"
	"cdn-service/handlers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize Config
	cfg := config.LoadConfig()

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
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("CDN Service Operational")
	})

	// Routing Group for Uploads
	// Secured by API Key
	api := app.Group("/api", handlers.CheckAPIKey(cfg))
	api.Post("/upload/avatar", handlers.Upload(cfg, "avatar"))
	api.Post("/upload/image", handlers.Upload(cfg, "image"))

	// Static File Serving (CDN) & Optimization
	// URLs: /avatar/filename.jpg?w=100
	// We use the same handler structure for both, but we could separate if needed.
	// NOTE: We replaced the direct Static middleware with our smart handler.

	app.Get("/:category/:filename", handlers.ServeOptimizedImage(cfg))

	// Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
