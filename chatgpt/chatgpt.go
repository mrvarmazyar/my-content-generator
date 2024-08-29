package chatgpt

import (
	"bytes"
	"encoding/json"
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

// GenerateArticle generates the title, description, tags, and article content based on selected keywords using the chat completion API
func GenerateArticle(keywords []string, apiKey string) (string, string, []string, string, error) {
	// Randomly select 2-3 keywords
	rand.Seed(time.Now().UnixNano())
	numKeywords := rand.Intn(2) + 2 // Select 2 or 3 keywords
	selectedKeywords := make([]string, numKeywords)

	for i := 0; i < numKeywords; i++ {
		index := rand.Intn(len(keywords))
		selectedKeywords[i] = keywords[index]
	}

	// Create a prompt for ChatGPT to generate the title, description, tags, and content
	keywordList := strings.Join(selectedKeywords, ", ")
	prompt := fmt.Sprintf(`Generate a detailed blog post with the following details:
1. Title covering these topics: %s.
2. A short description of the article.
3. Relevant tags as a comma-separated list.
4. Categories relevant to these topics as a comma-separated list.
5. The full article content.`, keywordList)

	requestBody := APIRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1500, // Adjust based on the length of the expected response
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // Increase the timeout duration
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", "", nil, "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", "", nil, "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", "", nil, "", fmt.Errorf("no content generated: %v", apiResponse)
	}

	// Assuming the response has the title, description, tags, categories, and content in sequence
	responseContent := apiResponse.Choices[0].Message.Content
	lines := strings.Split(responseContent, "\n")

	var title, description, content string
	var tags, categories []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimPrefix(line, "Title:")
			title = strings.TrimSpace(title)
			title = strings.Trim(title, `"'`) // Remove any leading/trailing quotes
		} else if strings.HasPrefix(line, "Description:") {
			description = strings.TrimPrefix(line, "Description:")
			description = strings.TrimSpace(description)
		} else if strings.HasPrefix(line, "Tags:") {
			tags = strings.Split(strings.TrimPrefix(line, "Tags:"), ", ")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		} else if strings.HasPrefix(line, "Categories:") {
			categories = strings.Split(strings.TrimPrefix(line, "Categories:"), ", ")
			for i, category := range categories {
				categories[i] = strings.TrimSpace(category)
			}
		} else {
			content += line + "\n"
		}
	}

	if title == "" || content == "" {
		return "", "", nil, "", fmt.Errorf("incomplete response from ChatGPT")
	}

	return title, description, tags, content, nil
}
