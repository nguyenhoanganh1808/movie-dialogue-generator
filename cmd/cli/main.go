// cmd/cli/main.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DialogueRequest struct {
	Scenario      string             `json:"scenario"`
	Characters    []CharacterRequest `json:"characters"`
	NumExchanges  int                `json:"numExchanges"`
	Style         string             `json:"style"`
	EmotionalTone string             `json:"emotionalTone"`
}

type CharacterRequest struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
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

func main() {
	// Define command line flags
	scenario := flag.String("scenario", "", "The scenario for the dialogue")
	style := flag.String("style", "", "The style of the dialogue (e.g., noir, comedy, drama)")
	tone := flag.String("tone", "", "The emotional tone of the dialogue")
	exchanges := flag.Int("exchanges", 5, "Number of dialogue exchanges to generate")
	apiURL := flag.String("api", "http://localhost:8080", "API base URL")
	
	flag.Parse()

	if *scenario == "" {
		fmt.Println("Error: Scenario is required")
		flag.Usage()
		os.Exit(1)
	}

	// Hard-coded characters for simplicity in this example
	// In a real CLI, you'd want to make this more flexible
	chars := []CharacterRequest{
		{
			Name:   "Detective Smith",
			Type:   "detective",
			Traits: []string{"cynical", "intelligent", "persistent"},
		},
		{
			Name:   "The Suspect",
			Type:   "villain",
			Traits: []string{"nervous", "calculating", "deceptive"},
		},
	}

	// Create request
	req := DialogueRequest{
		Scenario:      *scenario,
		Characters:    chars,
		NumExchanges:  *exchanges,
		Style:         *style,
		EmotionalTone: *tone,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error encoding request:", err)
		os.Exit(1)
	}

	// Send request to API
	resp, err := http.Post(*apiURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		fmt.Println("API Error:", string(body))
		os.Exit(1)
	}

	// Parse and display the dialogue
	var dialogue DialogueResponse
	err = json.Unmarshal(body, &dialogue)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		os.Exit(1)
	}

	// Display the dialogue
	fmt.Println("\n===== GENERATED DIALOGUE =====")
	fmt.Println("Scenario:", dialogue.Scenario)
	fmt.Println()

	for _, exchange := range dialogue.Exchanges {
		fmt.Printf("%s: %s\n\n", exchange.Character, exchange.Line)
	}
}