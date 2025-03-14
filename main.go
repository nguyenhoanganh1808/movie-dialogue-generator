// main.go
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/nguyenhoanganh1808/movie-dialogue-generator/api"
	"github.com/nguyenhoanganh1808/movie-dialogue-generator/db"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	db.InitDB()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for voice synthesis
	})

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
	log.Printf("Server started on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Movie Dialogue Generator API")
	})

	// API routes
	apiGroup := app.Group("/api")
	
	// Dialogue generation endpoint
	apiGroup.Post("/generate", api.GenerateDialogue)
	apiGroup.Post("/save-dialogue", api.SaveDialogue)
	apiGroup.Get("/saved-dialogues", api.GetSavedDialogues)
	
	// Character endpoints
	apiGroup.Get("/characters", api.GetCharacters)
	apiGroup.Post("/characters", api.CreateCharacter)
	
	// Reference dialogue endpoints
	apiGroup.Get("/references", api.GetReferenceDialogues)
	apiGroup.Post("/references", api.AddReferenceDialogue)
	
	// Voice synthesis endpoint (bonus feature)
	apiGroup.Post("/synthesize", api.SynthesizeVoice)
}