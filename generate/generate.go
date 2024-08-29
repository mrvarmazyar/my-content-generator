package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SanitizeFileName replaces or removes invalid characters for file names
func SanitizeFileName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove invalid characters (e.g., quotes, commas)
	re := regexp.MustCompile(`[^\w\-_]`)
	name = re.ReplaceAllString(name, "")

	return strings.ToLower(name)
}

// GenerateSlug is a public function to create a slug from a title
func GenerateSlug(title string) string {
	return SanitizeFileName(title)
}

// sanitizeTitle removes problematic characters from the title
func sanitizeTitle(title string) string {
	// Replace double quotes with single quotes
	title = strings.ReplaceAll(title, `"`, "'")

	return title
}

// extractShortDescription extracts the short description from the content
func extractShortDescription(content string) (string, string) {
	// Look for the "Short Description:" section and remove it
	if strings.Contains(content, "Short Description:") {
		parts := strings.SplitN(content, "Full Article:", 2)
		if len(parts) == 2 {
			description := strings.TrimSpace(strings.TrimPrefix(parts[0], "Short Description:"))
			content = strings.TrimSpace(parts[1])
			return description, content
		}
	}

	return "", content
}

// GenerateFrontMatter creates the front matter for the Markdown file
func GenerateFrontMatter(title, description string, tags []string) string {
	// Sanitize the title to remove any problematic characters
	title = sanitizeTitle(title)

	// Format the current date
	currentDate := time.Now().Format(time.RFC3339)

	// Generate the slug
	slug := SanitizeFileName(title)

	// Convert tags to a comma-separated string
	tagsStr := fmt.Sprintf(`["%s"]`, strings.Join(tags, `", "`))

	// Create the front matter
	frontMatter := fmt.Sprintf(`+++
draft = false
date = %s
title = "%s"
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

// EnsureDir ensures that a directory exists, creating it if necessary.
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}
	return nil
}

// SaveContent saves the generated content to a Markdown file in the 'generated' directory.
func SaveContent(title, content, description string, tags []string) error {
	// Sanitize the title to create a valid file name
	sanitizedTitle := SanitizeFileName(title)

	// Ensure the 'generated' directory exists
	err := EnsureDir("generated")
	if err != nil {
		return err
	}

	// Extract the short description from the content and update the description
	if description == "" {
		description, content = extractShortDescription(content)
	}

	filename := filepath.Join("generated", fmt.Sprintf("%s.md", sanitizedTitle))
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
