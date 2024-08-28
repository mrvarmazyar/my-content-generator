package chatgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// APIRequest represents the structure of the request to the OpenAI API for chat models
type APIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// Message represents a message to be sent in the chat
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIResponse represents the structure of the response from the OpenAI API
type APIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// GenerateContent generates a single blog post covering all the provided topics
func GenerateContent(topics []string, apiKey string) (string, error) {
	// Create a single prompt that asks to cover all topics
	prompt := fmt.Sprintf("I have a blog called 'Cloud Engineering Chronicles with Mohammad' where I automate posting articles. Please write a comprehensive article that covers the following topics: %s.",
		formatTopics(topics))

	requestBody := APIRequest{
		Model: "gpt-3.5-turbo", // or "gpt-4" if you have access
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1000, // Increase max tokens if you expect a longer post
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	var resp *http.Response
	for i := 0; i < 3; i++ { // Retry up to 3 times
		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
		if err != nil {
			return "", fmt.Errorf("failed to create new request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 30 * time.Second, // Increase the timeout duration
		}

		resp, err = client.Do(req)
		if err == nil {
			break // Exit loop if request is successful
		}
		if i < 2 { // Don't sleep after the last attempt
			time.Sleep(2 * time.Second) // Wait before retrying
		}
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", errors.New("no content generated")
	}

	return apiResponse.Choices[0].Message.Content, nil
}

// formatTopics formats the list of topics into a single string
func formatTopics(topics []string) string {
	return fmt.Sprintf("%s", topics)
}
