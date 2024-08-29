package generate

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// sanitizeFileName replaces or removes invalid characters for file names
func sanitizeFileName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove invalid characters (e.g., quotes, commas)
	re := regexp.MustCompile(`[^\w\-_]`)
	name = re.ReplaceAllString(name, "")

	return strings.ToLower(name)
}

// GenerateFrontMatter creates the front matter for the Markdown file
func GenerateFrontMatter(title, description string, tags []string) string {
	// Format the current date
	currentDate := time.Now().Format(time.RFC3339)

	// Generate the slug
	slug := sanitizeFileName(title)

	// Convert tags to a comma-separated string
	tagsStr := fmt.Sprintf(`["%s"]`, strings.Join(tags, `", "`))

	// Create the front matter
	frontMatter := fmt.Sprintf(`+++
draft = false
date = %s
title = %s
description = "%s"
slug = "%s"
authors = ["Mohammad Varmazyar"]
tags = %s
categories = %s
externalLink = ""
series = []
+++
`, currentDate, title, description, slug, tagsStr, tagsStr)

	return frontMatter
}

// SaveContent saves the generated content to a Markdown file with front matter.
func SaveContent(title, content string, description string, tags []string) error {
	// Sanitize the title to create a valid file name
	sanitizedTitle := sanitizeFileName(title)

	filename := fmt.Sprintf("%s.md", sanitizedTitle)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Generate the front matter
	frontMatter := GenerateFrontMatter(title, description, tags)

	// Write the front matter and content to the file
	_, err = file.WriteString(frontMatter + "\n" + content)
	if err != nil {
		return fmt.Errorf("failed to write content to file: %w", err)
	}

	fmt.Printf("Content saved to %s\n", filename)
	return nil
}
