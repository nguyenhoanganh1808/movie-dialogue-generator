// api/generate.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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

// HuggingFace API request structure
type HuggingFaceRequest struct {
	Inputs     string `json:"inputs"`
	Parameters struct {
		Temperature float64 `json:"temperature"`
		MaxNewTokens int    `json:"max_new_tokens"`
		ReturnFullText bool `json:"return_full_text"`
	} `json:"parameters"`
}

// HuggingFace API response structure
type HuggingFaceResponse struct {
	GeneratedText string `json:"generated_text"`
	Error         string `json:"error,omitempty"`
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
		systemPrompt := "You are a creative dialogue writer that specializes in creating authentic movie-like or anime-like dialogues. Create realistic exchanges between characters based on the described scenario and character traits."
	// Call HuggingFace API
	formattedPromt := formatLlamaPrompt(systemPrompt, prompt)

	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "HUGGINGFACE_API_KEY environment variable not set",
		})
	}
	modelID := os.Getenv("HUGGINGFACE_MODEL_ID")
	if modelID == "" {
		// Default to Llama 3.2 3B model if not specified
		modelID = "meta-llama/Meta-Llama-3.2-3B-Instruct"
	}


	// // Initialize OpenAI client
	// // Create request for HuggingFace Inference API
	hfReq := HuggingFaceRequest{
		Inputs: formattedPromt,
	}

	// // Set parameters
	hfReq.Parameters.Temperature = 0.7
	hfReq.Parameters.MaxNewTokens = 1024
	hfReq.Parameters.ReturnFullText = false

	// // Convert to JSON
	jsonData, err := json.Marshal(hfReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create request",
		})
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

		// Create request
		hfURL := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", modelID)
		request, err := http.NewRequest("POST", hfURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create request: %v", err),
			})
		}

	// 		// Set headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		// Send request
		resp, err := client.Do(request)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to connect to HuggingFace API: %v", err),
			})
		}
		defer resp.Body.Close()
	
		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to read response: %v", err),
			})
		}
	
		// Check for error status code
		if resp.StatusCode != http.StatusOK {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("HuggingFace API error (status %d): %s", resp.StatusCode, string(body)),
			})
		}
		log.Printf("Response: %s", string(body))
	
		// Parse response
		var hfResp []HuggingFaceResponse
		if err := json.Unmarshal(body, &hfResp); err != nil {
			// Try single response format
			var singleResp HuggingFaceResponse
			if err := json.Unmarshal(body, &singleResp); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to parse HuggingFace response: %v", err),
				})
			}
			hfResp = []HuggingFaceResponse{singleResp}
		}
	
		// Check for errors
		if len(hfResp) == 0 || hfResp[0].GeneratedText == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "No response generated",
			})
		}

		log.Println("Response:", hfResp[0].GeneratedText)
	
		// Process the AI response
		aiResponse := hfResp[0].GeneratedText
		dialogue := parseDialogueResponse(aiResponse, req.Characters)
	
		// Format the final response
		return c.JSON(DialogueResponse{
			Scenario:  req.Scenario,
			Exchanges: dialogue,
		})
	
}

// Format the prompt for Llama 3.2 with system and user roles
func formatLlamaPrompt(systemMessage, userMessage string) string {
	// The format for Llama 3.2 chat models
	prompt := fmt.Sprintf("<|system|>\n%s\n<|user|>\n%s\n<|assistant|>\n", systemMessage, userMessage)
	return prompt
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