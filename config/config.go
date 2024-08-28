package config

import (
	"os"
	"strings"
)

type Config struct {
	ChatGPTAPIKey string
	Keywords      []string
}

func LoadConfig() (*Config, error) {
	keywords := os.Getenv("KEYWORDS")
	if keywords == "" {
		keywords = "DevOps,Cloud Computing,SRE"
	}

	return &Config{
		ChatGPTAPIKey: os.Getenv("CHATGPT_API_KEY"),
		Keywords:      strings.Split(keywords, ","),
	}, nil
}
