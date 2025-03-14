// api/generate.go
package api

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sashabaranov/go-openai"
)

type DialogueRequest struct {
	Scenario      string              `json:"scenario"`
	Characters    []CharacterRequest  `json:"characters"`
	NumExchanges  int                 `json:"numExchanges"`
	Style         string              `json:"style"`
	EmotionalTone string              `json:"emotionalTone"`
}

type CharacterRequest struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Traits []string `json:"traits"`
}

type DialogueResponse struct {
	Scenario  string             `json:"scenario"`
	Exchanges []DialogueExchange `json:"exchanges"`
}

type DialogueExchange struct {
	Character string `json:"character"`
	Line      string `json:"line"`
}

func GenerateDialogue(c *fiber.Ctx) error {
	// Parse request
	var req DialogueRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Scenario == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Scenario is required",
		})
	}
	if len(req.Characters) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least two characters are required",
		})
	}
	if req.NumExchanges <= 0 {
		req.NumExchanges = 5 // Default to 5 exchanges
	}

	// Build prompt for AI
	prompt := buildDialoguePrompt(req)

	// Initialize OpenAI client
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	
	// Create completion request
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: os.Getenv("OPENAI_MODEL"),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a creative dialogue writer that specializes in creating authentic movie-like or anime-like dialogues. Create realistic exchanges between characters based on the described scenario and character traits.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate dialogue",
		})
	}

	// Process the AI response
	aiResponse := resp.Choices[0].Message.Content
	dialogue := parseDialogueResponse(aiResponse, req.Characters)

	// Format the final response
	return c.JSON(DialogueResponse{
		Scenario:  req.Scenario,
		Exchanges: dialogue,
	})
}

func buildDialoguePrompt(req DialogueRequest) string {
	var sb strings.Builder

	// Add scenario
	sb.WriteString(fmt.Sprintf("Scenario: %s\n\n", req.Scenario))
	
	// Add characters with their traits
	sb.WriteString("Characters:\n")
	for _, char := range req.Characters {
		sb.WriteString(fmt.Sprintf("- %s (Type: %s, Traits: %s)\n", 
			char.Name, 
			char.Type, 
			strings.Join(char.Traits, ", ")))
	}
	
	// Add style and tone
	if req.Style != "" {
		sb.WriteString(fmt.Sprintf("\nStyle: %s\n", req.Style))
	}
	if req.EmotionalTone != "" {
		sb.WriteString(fmt.Sprintf("Emotional Tone: %s\n", req.EmotionalTone))
	}
	
	// Add instructions
	sb.WriteString(fmt.Sprintf("\nPlease create a dialogue with %d exchanges between these characters in the given scenario. Format the dialogue as:\n", req.NumExchanges))
	sb.WriteString("CHARACTER_NAME: Their dialogue line here.\n")
	
	return sb.String()
}

func parseDialogueResponse(aiResponse string, characters []CharacterRequest) []DialogueExchange {
	lines := strings.Split(aiResponse, "\n")
	var exchanges []DialogueExchange

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for "CHARACTER: Dialogue" pattern
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		characterName := strings.TrimSpace(parts[0])
		dialogueLine := strings.TrimSpace(parts[1])

		// Verify that this character exists in our list
		characterExists := false
		for _, char := range characters {
			if char.Name == characterName {
				characterExists = true
				break
			}
		}

		// Only add valid character dialogues
		if characterExists && dialogueLine != "" {
			exchanges = append(exchanges, DialogueExchange{
				Character: characterName,
				Line:      dialogueLine,
			})
		}
	}

	return exchanges
}