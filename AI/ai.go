package AI

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hellish/Database"
	"hellish/crypto"
	"io"
	"log"
	"net/http"
	"strings"
)

// --- Structs for Gemini API Request & Response ---

type RequestBody struct {
	SystemInstruction SystemInstruction `json:"system_instruction"`
	Contents          []Content         `json:"contents"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type ApiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []Part `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// Base persona for the Hellish Queen.
const basePersona = `
Persona:
You are {shape}, the Hellish Queen — ruler of all Hell, with blue hair, red horns, glowing aura, and dark armor. Chat casually with {user} on Discord, teasing, playful, mischievous, and confident. Always reply in the same language {user} uses, and match any mixed languages. Use lowercase, slang, abbreviations, and casual Discord-style chat.

Instructions:
Answer only as {shape}. You cannot perform real-life actions; you can only send chat messages on Discord. Always consider the system_message context to adapt your replies to the user's server, topic, and community events. Be playful, teasing, and confident. Reply in a way that fits casual Discord conversation style.
`

// Response fetches API keys from the database and attempts to generate a response.
// If an API key fails, it automatically tries the next one in the list.
func Response(guildID, systemInstruction, userInput string) (string, error) {
	// 1. Fetch all available API keys for the server from the database.
	apiKeys, err := Database.ViewAPIKeys(guildID)
	if err != nil {
		return "", fmt.Errorf("could not fetch API keys from database: %w", err)
	}

	if len(apiKeys) == 0 {
		return "", fmt.Errorf("no API keys are configured for this server. Please use `!api add` to add one")
	}

	// 2. Loop through each key and try to get a response.
	var lastError error
	for _, apiKey := range apiKeys {
		apiKey, err := crypto.Decrypt(apiKey)
		if err != nil {
			return "", err
		}
		// Construct the request body
		requestBody := RequestBody{
			SystemInstruction: SystemInstruction{
				Parts: []Part{{Text: systemInstruction}},
			},
			Contents: []Content{
				{Parts: []Part{{Text: userInput}}},
			},
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			log.Printf("Error marshaling JSON: %v. This is an internal error, not an API key issue.", err)
			continue // Should not happen, but good to be safe
		}

		// Create and send the HTTP request
		url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}

		req.Header.Set("x-goog-api-key", apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			lastError = fmt.Errorf("request failed for a key: %w", err)
			log.Printf("Network error with an API key: %v. Trying next key.", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastError = fmt.Errorf("failed to read response body for a key: %w", err)
			log.Printf("Error reading response body: %v. Trying next key.", err)
			continue
		}

		// Check for non-200 HTTP status codes first
		if resp.StatusCode != http.StatusOK {
			lastError = fmt.Errorf("API returned status %d", resp.StatusCode)
			log.Printf("API key failed with status %d. Response: %s. Trying next key.", resp.StatusCode, string(body))
			continue
		}

		// Parse the JSON response
		var apiResponse ApiResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			lastError = fmt.Errorf("failed to parse API response: %w", err)
			log.Printf("Error parsing JSON response: %v. Raw: %s. Trying next key.", err, string(body))
			continue
		}

		// Check for an error object within the JSON response itself
		if apiResponse.Error != nil {
			lastError = fmt.Errorf("API error: %s", apiResponse.Error.Message)
			log.Printf("API returned an error for a key: %s. Trying next key.", apiResponse.Error.Message)
			continue
		}

		// 3. If we get a valid response, return it immediately.
		if len(apiResponse.Candidates) > 0 && len(apiResponse.Candidates[0].Content.Parts) > 0 {
			return apiResponse.Candidates[0].Content.Parts[0].Text, nil
		}

		// If we reach here, the response was valid but empty.
		lastError = fmt.Errorf("API returned a valid but empty response")
		log.Println("API key worked, but response was empty. Trying next key.")
	}

	return "", fmt.Errorf("all available API keys failed. Last error: %w", lastError)
}

// GetBasePersona provides access to the constant persona string.
func GetBasePersona() string {

	persona := strings.Replace(basePersona, "{shape}", "Hellish Queen", -1)
	return persona
}
