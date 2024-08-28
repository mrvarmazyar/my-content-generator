package main

import (
	"log"
	"my-content-generator/chatgpt"
	"my-content-generator/config"
	"my-content-generator/generate"
	"my-content-generator/publish"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Generate content for all keywords at once
	content, err := chatgpt.GenerateContent(cfg.Keywords, cfg.ChatGPTAPIKey)
	if err != nil {
		log.Fatalf("Error generating content: %v", err)
	}

	// Save the generated content as a post
	err = generate.SaveContent("Unified Article on Cloud Engineering", content)
	if err != nil {
		log.Fatalf("Error saving content: %v", err)
	}

	// Publish the content to GitHub
	err = publish.PublishToGitHub()
	if err != nil {
		log.Fatalf("Error publishing content: %v", err)
	}
}
