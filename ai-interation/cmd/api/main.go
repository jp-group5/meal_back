package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ai-interation/internal/config"
	"ai-interation/internal/httpapi"
	"ai-interation/internal/infra/openai"
	"ai-interation/internal/infra/repository"
	"ai-interation/internal/service"
)

func main() {
	cfg := config.Load()

	aiClient, err := openai.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewMockRepository()
	svc := service.NewService(aiClient, repo)

	r := gin.Default()
	httpapi.RegisterRoutes(r, httpapi.NewHandler(svc))

	log.Printf("server starting on %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}