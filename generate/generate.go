package generate

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// sanitizeFileName replaces or removes invalid characters for file names
func sanitizeFileName(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove invalid characters (e.g., quotes, commas)
	re := regexp.MustCompile(`[^\w\-_]`)
	name = re.ReplaceAllString(name, "")

	return name
}

// SaveContent saves the generated content to a Markdown file.
func SaveContent(title, content string) error {
	// Sanitize the title to create a valid file name
	sanitizedTitle := sanitizeFileName(title)

	filename := fmt.Sprintf("%s.md", sanitizedTitle)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content to file: %w", err)
	}

	fmt.Printf("Content saved to %s\n", filename)
	return nil
}
