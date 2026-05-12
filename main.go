package main

import (
	"log"
	"os"

	"meal_back/handlers"
	"meal_back/middlewares"
	"meal_back/models"
	"meal_back/stores"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	jwtSecret := os.Getenv("JWT_SECRET")

	if dsn == "" {
		log.Fatal("Missing environment variable DB_DSN")
	}
	if jwtSecret == "" {
		log.Fatal("Missing environment variable JWT_SECRET")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 迁移顺序：先用户主表，再会话/资料/业务记录表。
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.UserProfile{},
		&models.MealRecord{},
		&models.ActivityRecord{},
	); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	tokenBlacklist := stores.NewTokenBlacklistStore()
	authHandler := handlers.NewAuthHandler(db, jwtSecret, tokenBlacklist)
	nutritionHandler := handlers.NewNutritionHandler(db)

	r := gin.Default()
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/register", authHandler.Register)
		apiV1.POST("/login", authHandler.Login)
		apiV1.POST("/refresh", authHandler.RefreshToken)

		authed := apiV1.Group("")
		authed.Use(middlewares.AuthMiddleware(db, jwtSecret, tokenBlacklist))
		authed.GET("/private/me", authHandler.Me)
		authed.POST("/private/logout", authHandler.Logout)
		authed.PUT("/private/me/profile", authHandler.UpsertProfile)
		authed.POST("/private/me/profile", authHandler.UpsertProfile)
		authed.PUT("/users/me/profile", authHandler.UpsertProfile)
		authed.POST("/users/me/profile", authHandler.UpsertProfile)
		authed.PUT("/users/me/preferences", nutritionHandler.UpsertPreferences)
		authed.PUT("/private/me/preferences", nutritionHandler.UpsertPreferences)

		authed.GET("/meals", nutritionHandler.GetMealsByDate)
		authed.POST("/meals", nutritionHandler.CreateMeal)
		authed.PUT("/meals/:id", nutritionHandler.UpdateMeal)
		authed.DELETE("/meals/:id", nutritionHandler.DeleteMeal)

		authed.GET("/activities", nutritionHandler.GetActivitiesByDate)
		authed.POST("/activities", nutritionHandler.CreateActivity)
		authed.PUT("/activities/:id", nutritionHandler.UpdateActivity)
		authed.DELETE("/activities/:id", nutritionHandler.DeleteActivity)

		authed.POST("/recommendations", nutritionHandler.GetRecommendation)
		authed.GET("/recommendations/prompt", nutritionHandler.PreviewRecommendationPrompt)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
