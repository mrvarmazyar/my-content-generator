package chatgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
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

// GenerateArticle generates both the title (question) and the article content based on selected keywords using the chat completion API
func GenerateArticle(keywords []string, apiKey string) (string, string, error) {
	// Randomly select 2-3 keywords
	rand.Seed(time.Now().UnixNano())
	numKeywords := rand.Intn(2) + 2 // Select 2 or 3 keywords
	selectedKeywords := make([]string, numKeywords)

	for i := 0; i < numKeywords; i++ {
		index := rand.Intn(len(keywords))
		selectedKeywords[i] = keywords[index]
	}

	// Create a prompt for ChatGPT to generate both a question and the content
	keywordList := strings.Join(selectedKeywords, ", ")
	prompt := fmt.Sprintf("Generate a blog post title and content that covers the following topics: %s. First, provide a title, then write the article.", keywordList)

	requestBody := APIRequest{
		Model: "gpt-3.5-turbo", // or "gpt-4" if you have access
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1500, // Increase max tokens to handle both title and content
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	var resp *http.Response
	for i := 0; i < 3; i++ { // Retry up to 3 times
		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
		if err != nil {
			return "", "", fmt.Errorf("failed to create new request: %w", err)
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
		return "", "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", "", errors.New("no content generated")
	}

	// Assuming the response has the title and content in the generated text
	responseContent := apiResponse.Choices[0].Message.Content
	splitContent := strings.SplitN(responseContent, "\n", 2)
	if len(splitContent) < 2 {
		return "", "", errors.New("response format unexpected")
	}

	title := strings.TrimSpace(splitContent[0])   // The first line is the title
	content := strings.TrimSpace(splitContent[1]) // The rest is the content

	return title, content, nil
}
