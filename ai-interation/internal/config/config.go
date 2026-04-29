package config

import "os"

type Config struct {
	Port         string
	OpenAIAPIKey string
	OpenAIModel  string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = ":" + port
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	return Config{
		Port:         port,
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  model,
	}
}