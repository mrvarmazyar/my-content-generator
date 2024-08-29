package main

import (
	"fmt"
	"log"
	"my-content-generator/chatgpt"
	"my-content-generator/config"
	"my-content-generator/db"
	"my-content-generator/generate"
	"my-content-generator/publish"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Load keywords from JSON file
	keywords, err := db.LoadKeywords("db/keywords.json")
	if err != nil {
		log.Fatalf("Error loading keywords: %v", err)
	}

	// Generate the title, description, tags, and article content using the keywords and ChatGPT
	title, description, tags, content, err := chatgpt.GenerateArticle(keywords, cfg.ChatGPTAPIKey)
	if err != nil {
		log.Fatalf("Error generating article: %v", err)
	}
	fmt.Println("Generated Title:", title)
	fmt.Println("Generated Description:", description)
	fmt.Println("Generated Tags:", tags)
	fmt.Println("Generated Content:", content)

	// Save the generated content as a post with front matter
	err = generate.SaveContent(title, content, description, tags)
	if err != nil {
		log.Fatalf("Error saving content: %v", err)
	}

	// Publish the content to GitHub
	err = publish.PublishToGitHub()
	if err != nil {
		log.Fatalf("Error publishing content: %v", err)
	}
}
