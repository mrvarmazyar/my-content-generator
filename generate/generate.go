package generate

import (
	"fmt"
	"log"
	"os"
)

// SaveContent saves the generated content to a Markdown file.
func SaveContent(title, content string) error {
	filename := fmt.Sprintf("%s.md", title)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content to file: %w", err)
	}

	log.Printf("Content saved to %s", filename)
	return nil
}
