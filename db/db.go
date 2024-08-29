package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type KeywordsDB struct {
	Keywords []string `json:"keywords"`
}

// LoadKeywords loads keywords from a JSON file
func LoadKeywords(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var db KeywordsDB
	if err := json.Unmarshal(byteValue, &db); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return db.Keywords, nil
}
