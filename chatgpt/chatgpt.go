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
	prompt := fmt.Sprintf(`Generate a comprehensive and detailed blog post with the following details:
1. Title covering these topics: %s.
2. A short description of the article.
3. Relevant tags as a comma-separated list.
4. The full article content. The content should be in-depth, covering multiple aspects of the topics, including challenges, best practices, and real-world examples. It should be designed to take around 10 minutes to read, including sections that discuss the challenges faced when working with these technologies and how to overcome them.`, keywordList)

	requestBody := APIRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "system", Content: "You are a highly skilled technical writer."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 3000, // Increase to allow for more detailed content
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	var resp *http.Response
	for i := 0; i < 3; i++ { // Retry up to 3 times
		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
		if err != nil {
			return "", "", nil, "", fmt.Errorf("failed to create new request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 60 * time.Second, // Increase the timeout duration
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
		return "", "", nil, "", errors.New("no content generated")
	}

	// Assuming the response has the title, description, tags, and content in sequence
	responseContent := apiResponse.Choices[0].Message.Content
	lines := strings.Split(responseContent, "\n")

	var title, description, content string
	var tags []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimPrefix(line, "Title:")
			title = strings.TrimSpace(title)
		} else if strings.HasPrefix(line, "Description:") {
			description = strings.TrimPrefix(line, "Description:")
			description = strings.TrimSpace(description)
		} else if strings.HasPrefix(line, "Tags:") {
			tags = strings.Split(strings.TrimPrefix(line, "Tags:"), ", ")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		} else {
			content += line + "\n"
		}
	}

	if title == "" || content == "" {
		return "", "", nil, "", errors.New("incomplete response from ChatGPT")
	}

	return title, description, tags, content, nil
}
