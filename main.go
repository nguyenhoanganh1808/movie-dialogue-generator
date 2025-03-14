// main.go
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// Create Fiber app
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	setupRoutes(app)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Movie Dialogue Generator API")
	})

	// API routes
	api := app.Group("/api")
	
	// Dialogue generation endpoint
	api.Post("/generate", generateDialogue)
	
	// Character endpoints
	api.Get("/characters", getCharacters)
	api.Post("/characters", createCharacter)
	
	// Reference dialogue endpoints
	api.Get("/references", getReferenceDialogues)
	api.Post("/references", addReferenceDialogue)
}

