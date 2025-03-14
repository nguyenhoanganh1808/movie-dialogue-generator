// api/handlers.go
package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/nguyenhoanganh1808/movie-dialogue-generator/db"
)

// Character handlers
func GetCharacters(c *fiber.Ctx) error {
	characters, err := db.GetCharacters()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch characters",
		})
	}

	return c.JSON(characters)
}

func CreateCharacter(c *fiber.Ctx) error {
	var character db.Character
	if err := c.BodyParser(&character); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	id, err := db.CreateCharacter(character)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create character",
		})
	}

	character.ID = id
	return c.Status(fiber.StatusCreated).JSON(character)
}

// Reference dialogue handlers
func GetReferenceDialogues(c *fiber.Ctx) error {
	tag := c.Query("tag")
	
	dialogues, err := db.GetReferenceDialogues(tag)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch reference dialogues",
		})
	}

	return c.JSON(dialogues)
}

func AddReferenceDialogue(c *fiber.Ctx) error {
	var dialogue db.ReferenceDialogue
	if err := c.BodyParser(&dialogue); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	id, err := db.AddReferenceDialogue(dialogue)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add reference dialogue",
		})
	}

	dialogue.ID = id
	return c.Status(fiber.StatusCreated).JSON(dialogue)
}

// Generated dialogue handlers
func SaveDialogue(c *fiber.Ctx) error {
	var req struct {
		Scenario   string          `json:"scenario"`
		Characters []CharacterRequest `json:"characters"`
		Exchanges  []DialogueExchange `json:"exchanges"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert to JSON for storage
	charactersJSON, err := json.Marshal(req.Characters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize characters",
		})
	}

	exchangesJSON, err := json.Marshal(req.Exchanges)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize exchanges",
		})
	}

	id, err := db.SaveGeneratedDialogue(req.Scenario, charactersJSON, exchangesJSON)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save dialogue",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
		"message": "Dialogue saved successfully",
	})
}

func GetSavedDialogues(c *fiber.Ctx) error {
	dialogues, err := db.GetGeneratedDialogues()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch saved dialogues",
		})
	}

	return c.JSON(dialogues)
}