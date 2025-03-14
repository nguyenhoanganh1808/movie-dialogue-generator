package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

type VoiceRequest struct {
	Character string `json:"character"`
	Text      string `json:"text"`
	VoiceID   string `json:"voiceId"` // Optional: specific voice ID to use
}

type ElevenLabsRequest struct {
	Text      string    `json:"text"`
	ModelID   string    `json:"model_id"`
	VoiceSettings VoiceSettings `json:"voice_settings"`
}

type VoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
}

func SynthesizeVoice(c *fiber.Ctx) error {
	var req VoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Text is required",
		})
	}

	// Get voice ID (either from request or based on character)
	voiceID := req.VoiceID
	if voiceID == "" {
		// Map character to a default voice ID
		// This would be expanded with more sophisticated mapping in production
		voiceMap := map[string]string{
			"hero":      "21m00Tcm4TlvDq8ikWAM", // Example voice IDs
			"villain":   "AZnzlk1XvdvUeBnXmlld",
			"sidekick":  "EXAVITQu4vr4xnSDxMaL",
			"detective": "MF3mGyEYCl7XYWbV9V6O",
			"default":   "EXAVITQu4vr4xnSDxMaL",
		}

		_, ok := voiceMap[req.Character]
		if !ok {
			voiceID = voiceMap["default"]
		}
	}

	// Prepare request to ElevenLabs API
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Voice synthesis API key not configured",
		})
	}

	elevenLabsReq := ElevenLabsRequest{
		Text:    req.Text,
		ModelID: "eleven_monolingual_v1",
		VoiceSettings: VoiceSettings{
			Stability:       0.75,
			SimilarityBoost: 0.75,
		},
	}

	jsonData, err := json.Marshal(elevenLabsReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize request",
		})
	}

	// Make request to ElevenLabs API
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create HTTP request",
		})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to make voice synthesis request",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Voice synthesis API error",
			"response": string(bodyBytes),
		})
	}

	// Set appropriate headers for audio file
	c.Set("Content-Type", "audio/mpeg")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_dialogue.mp3\"", req.Character))

	// Stream the audio data directly to the client
	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read audio response",
		})
	}

	return c.Send(audio)
}