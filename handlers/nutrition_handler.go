package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"meal_back/models"
	"meal_back/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	codeNutritionInvalidParam = 11001
	codeNutritionNotFound     = 11002
	codeNutritionDBError      = 11003
)

type NutritionHandler struct {
	db *gorm.DB
}

func NewNutritionHandler(db *gorm.DB) *NutritionHandler {
	return &NutritionHandler{db: db}
}

type upsertPreferencesRequest struct {
	Allergies             []string `json:"allergies"`
	DietaryPreferences    []string `json:"dietary_preferences"`
	DietaryPreferencesAlt []string `json:"dietaryPreferences"`
}

type mealRequest struct {
	Date string `json:"date"`

	Type string `json:"type"`

	Content string `json:"content"`
	Name    string `json:"name"`

	Calories *int `json:"calories"`

	Protein *float64 `json:"protein"`
	Carbs   *float64 `json:"carbs"`
	Fat     *float64 `json:"fat"`

	MealType string `json:"mealType"`

	Source           string `json:"source"`
	RecommendationID string `json:"recommendationId"`
}

type activityRequest struct {
	Title string `json:"title"`
	Date  string `json:"date"`

	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	StartTimeAlt string `json:"start_time"`
	EndTimeAlt   string `json:"end_time"`

	Intensity string `json:"intensity"`
}

type recommendationRequest struct {
	Date string `json:"date" binding:"required"`
}

func (h *NutritionHandler) UpsertPreferences(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	var req upsertPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Invalid request payload.")
		return
	}

	allergies := normalizeStringList(req.Allergies)
	dietaryPreferences := normalizeStringList(req.DietaryPreferences)
	if len(dietaryPreferences) == 0 {
		dietaryPreferences = normalizeStringList(req.DietaryPreferencesAlt)
	}

	var profile models.UserProfile
	if err := h.db.Where("user_id = ?", userID).FirstOrCreate(&profile, models.UserProfile{UserID: userID}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to initialize user profile.")
		return
	}

	profile.Allergies = allergies
	profile.DietaryPreferences = dietaryPreferences
	if err := h.db.Save(&profile).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to save user preferences.")
		return
	}

	var user models.User
	if err := h.db.Select("id", "username", "email", "created_at", "updated_at").First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query user.")
		return
	}
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query user profile.")
		return
	}

	response.Success(c, http.StatusOK, buildUserPayload(user, &profile))
}

func (h *NutritionHandler) GetMealsByDate(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	dateRaw := strings.TrimSpace(c.Query("date"))
	day, err := parseDate(dateRaw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	var records []models.MealRecord
	if err := h.db.
		Where("user_id = ? AND date = ?", userID, day.Format(dateLayout)).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query meal records.")
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, toMealPayload(record))
	}
	response.Success(c, http.StatusOK, items)
}

func (h *NutritionHandler) CreateMeal(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	var req mealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Invalid request payload.")
		return
	}

	record, err := buildMealRecord(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to create meal record.")
		return
	}

	response.Success(c, http.StatusCreated, toMealPayload(record))
}

func (h *NutritionHandler) UpdateMeal(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	mealID, err := parseUintParam(c.Param("id"), "meal id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	var record models.MealRecord
	if err := h.db.Where("id = ? AND user_id = ?", mealID, userID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, codeNutritionNotFound, "Meal record not found.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query meal record.")
		return
	}

	var req mealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Invalid request payload.")
		return
	}

	updates, err := buildMealUpdates(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}
	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Provide at least one updatable field.")
		return
	}

	if err := h.db.Model(&record).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to update meal record.")
		return
	}

	if err := h.db.First(&record, record.ID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query updated meal record.")
		return
	}
	response.Success(c, http.StatusOK, toMealPayload(record))
}

func (h *NutritionHandler) DeleteMeal(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	mealID, err := parseUintParam(c.Param("id"), "meal id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", mealID, userID).Delete(&models.MealRecord{})
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to delete meal record.")
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, codeNutritionNotFound, "Meal record not found.")
		return
	}

	response.Success(c, http.StatusOK, nil)
}

func (h *NutritionHandler) GetActivitiesByDate(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	dateRaw := strings.TrimSpace(c.Query("date"))
	day, err := parseDate(dateRaw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	var records []models.ActivityRecord
	if err := h.db.
		Where("user_id = ? AND date = ?", userID, day.Format(dateLayout)).
		Order("start_time ASC").
		Find(&records).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query activity records.")
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, toActivityPayload(record))
	}
	response.Success(c, http.StatusOK, items)
}

func (h *NutritionHandler) CreateActivity(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Invalid request payload.")
		return
	}

	record, err := buildActivityRecord(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to create activity record.")
		return
	}
	response.Success(c, http.StatusCreated, toActivityPayload(record))
}

func (h *NutritionHandler) UpdateActivity(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	activityID, err := parseUintParam(c.Param("id"), "activity id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	var record models.ActivityRecord
	if err := h.db.Where("id = ? AND user_id = ?", activityID, userID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, codeNutritionNotFound, "Activity record not found.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query activity record.")
		return
	}

	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Invalid request payload.")
		return
	}

	updates, err := buildActivityUpdates(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}
	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "Provide at least one updatable field.")
		return
	}

	if err := h.db.Model(&record).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to update activity record.")
		return
	}
	if err := h.db.First(&record, record.ID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to query updated activity record.")
		return
	}

	response.Success(c, http.StatusOK, toActivityPayload(record))
}

func (h *NutritionHandler) DeleteActivity(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	activityID, err := parseUintParam(c.Param("id"), "activity id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", activityID, userID).Delete(&models.ActivityRecord{})
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to delete activity record.")
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, codeNutritionNotFound, "Activity record not found.")
		return
	}

	response.Success(c, http.StatusOK, nil)
}

func (h *NutritionHandler) GetRecommendation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	var req recommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, "date is required in YYYY-MM-DD format.")
		return
	}

	targetDate, err := parseDate(req.Date)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	promptData, err := h.buildRecommendationPrompt(userID, targetDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to build recommendation context.")
		return
	}

	recommendation := buildRecommendationFromPrompt(targetDate, promptData)
	response.Success(c, http.StatusOK, recommendation)
}

func (h *NutritionHandler) PreviewRecommendationPrompt(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	dateRaw := strings.TrimSpace(c.Query("date"))
	targetDate, err := parseDate(dateRaw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, codeNutritionInvalidParam, err.Error())
		return
	}

	promptData, err := h.buildRecommendationPrompt(userID, targetDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeNutritionDBError, "Failed to build recommendation context.")
		return
	}
	response.Success(c, http.StatusOK, promptData)
}

type recommendationPromptData struct {
	Metadata recommendationMetadata `json:"metadata"`
	User     recommendationUser     `json:"user"`
	Context  recommendationContext  `json:"context"`
}

type recommendationMetadata struct {
	GeneratedAt string `json:"generated_at"`
	TargetDate  string `json:"target_date"`
	TimeZone    string `json:"time_zone"`
}

type recommendationUser struct {
	UserID              uint      `json:"user_id"`
	Username            string    `json:"username"`
	HeightCM            *float64  `json:"height_cm,omitempty"`
	WeightKG            *float64  `json:"weight_kg,omitempty"`
	FitnessGoal         string    `json:"fitness_goal,omitempty"`
	TrainingExperience  []string  `json:"training_experience"`
	MonthlyFoodBudget   *int64    `json:"monthly_food_budget,omitempty"`
	Allergies           []string  `json:"allergies"`
	DietaryPreferences  []string  `json:"dietary_preferences"`
	ExerciseExperience  string    `json:"exercise_experience,omitempty"`
	ProfileLastUpdateAt time.Time `json:"profile_last_update_at"`
}

type recommendationContext struct {
	RecentMeals           []recommendationMeal     `json:"recent_meals"`
	WeeklyActivities      []recommendationActivity `json:"weekly_activities"`
	MealStats             mealStats                `json:"meal_stats"`
	ActivityIntensityStat map[string]int           `json:"activity_intensity_stats"`
}

type recommendationMeal struct {
	Date     string   `json:"date"`
	Type     string   `json:"type"`
	Content  string   `json:"content"`
	Calories *int     `json:"calories,omitempty"`
	Protein  *float64 `json:"protein,omitempty"`
	Carbs    *float64 `json:"carbs,omitempty"`
	Fat      *float64 `json:"fat,omitempty"`
}

type recommendationActivity struct {
	Date      string `json:"date"`
	Title     string `json:"title"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Intensity string `json:"intensity,omitempty"`
}

type mealStats struct {
	DaysWithMeals         int `json:"days_with_meals"`
	TotalMealRecords      int `json:"total_meal_records"`
	AverageDailyCalories  int `json:"average_daily_calories,omitempty"`
	RecordsWithCalories   int `json:"records_with_calories"`
	CaloriesWindowInDays  int `json:"calories_window_in_days"`
	RecentMealsWindowDays int `json:"recent_meals_window_days"`
}

func (h *NutritionHandler) buildRecommendationPrompt(userID uint, targetDate time.Time) (*recommendationPromptData, error) {
	var user models.User
	if err := h.db.Select("id", "username", "email").First(&user, userID).Error; err != nil {
		return nil, err
	}

	var profile models.UserProfile
	if err := h.db.Where("user_id = ?", userID).FirstOrCreate(&models.UserProfile{UserID: userID}).Error; err != nil {
		return nil, err
	}
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}

	recentStart := targetDate.AddDate(0, 0, -6)
	var meals []models.MealRecord
	if err := h.db.
		Where("user_id = ? AND date >= ? AND date <= ?", userID, recentStart.Format(dateLayout), targetDate.Format(dateLayout)).
		Order("date DESC, created_at DESC").
		Find(&meals).Error; err != nil {
		return nil, err
	}

	weekStart, weekEnd := weekRange(targetDate)
	var activities []models.ActivityRecord
	if err := h.db.
		Where("user_id = ? AND date >= ? AND date <= ?", userID, weekStart.Format(dateLayout), weekEnd.Format(dateLayout)).
		Order("date ASC, start_time ASC").
		Find(&activities).Error; err != nil {
		return nil, err
	}

	trainings := parseTrainingExperienceText(profile.ExerciseExperience)
	promptMeals := make([]recommendationMeal, 0, len(meals))
	dailyCalories := make(map[string]int)
	intensityStats := map[string]int{"low": 0, "medium": 0, "high": 0, "unknown": 0}
	calorieRecords := 0
	totalCalories := 0

	for _, meal := range meals {
		payload := recommendationMeal{
			Date:     meal.Date.Format(dateLayout),
			Type:     strings.ToLower(strings.TrimSpace(meal.Type)),
			Content:  meal.Content,
			Calories: meal.Calories,
			Protein:  meal.Protein,
			Carbs:    meal.Carbs,
			Fat:      meal.Fat,
		}
		promptMeals = append(promptMeals, payload)
		if meal.Calories != nil && *meal.Calories > 0 {
			dailyCalories[payload.Date] += *meal.Calories
			totalCalories += *meal.Calories
			calorieRecords++
		}
	}

	promptActivities := make([]recommendationActivity, 0, len(activities))
	for _, activity := range activities {
		intensity := normalizeIntensity(activity.Intensity)
		if intensity == "" {
			intensity = "unknown"
		}
		intensityStats[intensity]++
		promptActivities = append(promptActivities, recommendationActivity{
			Date:      activity.Date.Format(dateLayout),
			Title:     activity.Title,
			StartTime: activity.StartTime,
			EndTime:   activity.EndTime,
			Intensity: intensity,
		})
	}

	daysWithMeals := len(dailyCalories)
	avgDailyCalories := 0
	if daysWithMeals > 0 {
		avgDailyCalories = totalCalories / daysWithMeals
	}

	return &recommendationPromptData{
		Metadata: recommendationMetadata{
			GeneratedAt: time.Now().Format(time.RFC3339),
			TargetDate:  targetDate.Format(dateLayout),
			TimeZone:    "Asia/Tokyo",
		},
		User: recommendationUser{
			UserID:              user.ID,
			Username:            user.Username,
			HeightCM:            profile.HeightCM,
			WeightKG:            profile.WeightKG,
			FitnessGoal:         profile.FitnessGoal,
			TrainingExperience:  trainings,
			MonthlyFoodBudget:   profile.MonthlyDietBudget,
			Allergies:           ensureNonNilSlice(profile.Allergies),
			DietaryPreferences:  ensureNonNilSlice(profile.DietaryPreferences),
			ExerciseExperience:  profile.ExerciseExperience,
			ProfileLastUpdateAt: profile.UpdatedAt,
		},
		Context: recommendationContext{
			RecentMeals:      promptMeals,
			WeeklyActivities: promptActivities,
			MealStats: mealStats{
				DaysWithMeals:         daysWithMeals,
				TotalMealRecords:      len(meals),
				AverageDailyCalories:  avgDailyCalories,
				RecordsWithCalories:   calorieRecords,
				CaloriesWindowInDays:  7,
				RecentMealsWindowDays: 7,
			},
			ActivityIntensityStat: intensityStats,
		},
	}, nil
}

func buildRecommendationFromPrompt(targetDate time.Time, promptData *recommendationPromptData) gin.H {
	goal := promptData.User.FitnessGoal
	avgCalories := promptData.Context.MealStats.AverageDailyCalories
	activity := promptData.Context.ActivityIntensityStat
	highIntensityCount := activity["high"]

	baseTitle := "Daily Nutrition Recommendation"
	baseReason := "These options are generated from your profile, recent meal records, and this week's activities."

	switch goal {
	case "lose_weight":
		baseTitle = "Fat Loss Nutrition Options"
		baseReason = "Your goal is fat loss, so these options prioritize protein, fiber, and controlled refined carbs."
	case "build_muscle":
		baseTitle = "Muscle Gain Nutrition Options"
		baseReason = "Your goal is muscle gain, so these options increase quality protein and complex carbs around training."
	case "maintain_shape":
		baseTitle = "Body Maintenance Nutrition Options"
		baseReason = "Your goal is maintenance, so these options keep intake stable and meals consistent."
	}

	if highIntensityCount >= 2 {
		baseReason += " You had multiple high-intensity sessions this week, so moderate carb replenishment is included."
	}
	if avgCalories > 0 && avgCalories < 1400 {
		baseReason += " Recent average calories are low, so options avoid prolonged under-fueling."
	}

	options := []gin.H{
		{
			"optionId": "plan-a",
			"title":    "Option A - Balanced Performance",
			"reason":   baseReason + " This option balances satiety and recovery.",
			"suggestedMeals": []gin.H{
				{"type": "breakfast", "content": "Greek yogurt oatmeal bowl + boiled egg + berries"},
				{"type": "lunch", "content": "Chicken and brown rice salad bowl + avocado"},
				{"type": "dinner", "content": "Steamed fish + broccoli + sweet potato"},
			},
		},
		{
			"optionId": "plan-b",
			"title":    "Option B - Higher Protein",
			"reason":   baseReason + " This option increases protein density across all meals.",
			"suggestedMeals": []gin.H{
				{"type": "breakfast", "content": "Egg white omelet + cottage cheese + apple"},
				{"type": "lunch", "content": "Turkey quinoa bowl + mixed greens"},
				{"type": "dinner", "content": "Grilled salmon + asparagus + lentils"},
			},
		},
		{
			"optionId": "plan-c",
			"title":    "Option C - Quick Prep",
			"reason":   baseReason + " This option is designed for lower preparation time.",
			"suggestedMeals": []gin.H{
				{"type": "breakfast", "content": "Protein smoothie + banana + peanut-free granola"},
				{"type": "lunch", "content": "Tuna whole-grain wrap + side salad"},
				{"type": "dinner", "content": "Tofu stir-fry + microwave brown rice + vegetables"},
			},
		},
	}

	if len(promptData.User.DietaryPreferences) > 0 {
		pref := promptData.User.DietaryPreferences[0]
		for idx := range options {
			meals, ok := options[idx]["suggestedMeals"].([]gin.H)
			if !ok || len(meals) == 0 {
				continue
			}
			meals[0]["content"] = fmt.Sprintf("Preference-based (%s): soy milk + whole-grain toast + mixed nuts", pref)
			options[idx]["suggestedMeals"] = meals
		}
	}

	primary := options[0]
	recommendationID := fmt.Sprintf("rec-%d", time.Now().Unix())

	return gin.H{
		"date":               targetDate.Format(dateLayout),
		"title":              baseTitle,
		"default_choice":     primary["title"],
		"reason":             primary["reason"],
		"suggestedMeals":     primary["suggestedMeals"],
		"recommendationId":   recommendationID,
		"choice_count":       len(options),
		"choices":            options,
		"selection_guidance": "Choose one option based on your schedule, appetite, and meal preparation time.",
		"prompt_json":        promptData,
		"prompt_version":     "v2",
	}
}

func toMealPayload(record models.MealRecord) gin.H {
	return gin.H{
		"id":               strconv.FormatUint(uint64(record.ID), 10),
		"date":             record.Date.Format(dateLayout),
		"type":             strings.ToLower(strings.TrimSpace(record.Type)),
		"content":          record.Content,
		"calories":         record.Calories,
		"protein":          record.Protein,
		"carbs":            record.Carbs,
		"fat":              record.Fat,
		"source":           record.Source,
		"recommendationId": record.RecommendationID,
	}
}

func toActivityPayload(record models.ActivityRecord) gin.H {
	return gin.H{
		"id":        strconv.FormatUint(uint64(record.ID), 10),
		"title":     record.Title,
		"date":      record.Date.Format(dateLayout),
		"startTime": record.StartTime,
		"endTime":   record.EndTime,
		"intensity": normalizeIntensity(record.Intensity),
	}
}

func buildMealRecord(userID uint, req mealRequest) (models.MealRecord, error) {
	day, err := parseDate(req.Date)
	if err != nil {
		return models.MealRecord{}, err
	}
	mealType, err := normalizeMealType(firstNotEmpty(req.Type, req.MealType))
	if err != nil {
		return models.MealRecord{}, err
	}
	content := strings.TrimSpace(firstNotEmpty(req.Content, req.Name))
	if content == "" {
		return models.MealRecord{}, errors.New("content is required.")
	}
	if req.Calories != nil && *req.Calories < 0 {
		return models.MealRecord{}, errors.New("calories cannot be negative.")
	}

	record := models.MealRecord{
		UserID:           userID,
		Date:             day,
		Type:             mealType,
		Content:          content,
		Calories:         req.Calories,
		Protein:          req.Protein,
		Carbs:            req.Carbs,
		Fat:              req.Fat,
		Source:           strings.TrimSpace(req.Source),
		RecommendationID: strings.TrimSpace(req.RecommendationID),
	}
	return record, nil
}

func buildMealUpdates(req mealRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{}

	if strings.TrimSpace(req.Date) != "" {
		day, err := parseDate(req.Date)
		if err != nil {
			return nil, err
		}
		updates["date"] = day
	}

	if strings.TrimSpace(req.Type) != "" || strings.TrimSpace(req.MealType) != "" {
		mealType, err := normalizeMealType(firstNotEmpty(req.Type, req.MealType))
		if err != nil {
			return nil, err
		}
		updates["type"] = mealType
	}

	if strings.TrimSpace(req.Content) != "" || strings.TrimSpace(req.Name) != "" {
		content := strings.TrimSpace(firstNotEmpty(req.Content, req.Name))
		updates["content"] = content
	}

	if req.Calories != nil {
		if *req.Calories < 0 {
			return nil, errors.New("calories cannot be negative.")
		}
		updates["calories"] = *req.Calories
	}
	if req.Protein != nil {
		updates["protein"] = *req.Protein
	}
	if req.Carbs != nil {
		updates["carbs"] = *req.Carbs
	}
	if req.Fat != nil {
		updates["fat"] = *req.Fat
	}
	if strings.TrimSpace(req.Source) != "" {
		updates["source"] = strings.TrimSpace(req.Source)
	}
	if strings.TrimSpace(req.RecommendationID) != "" {
		updates["recommendation_id"] = strings.TrimSpace(req.RecommendationID)
	}

	return updates, nil
}

func buildActivityRecord(userID uint, req activityRequest) (models.ActivityRecord, error) {
	day, err := parseDate(req.Date)
	if err != nil {
		return models.ActivityRecord{}, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return models.ActivityRecord{}, errors.New("title is required.")
	}

	start := strings.TrimSpace(firstNotEmpty(req.StartTime, req.StartTimeAlt))
	end := strings.TrimSpace(firstNotEmpty(req.EndTime, req.EndTimeAlt))

	if err := validateClock(start); err != nil {
		return models.ActivityRecord{}, fmt.Errorf("startTime %w", err)
	}
	if err := validateClock(end); err != nil {
		return models.ActivityRecord{}, fmt.Errorf("endTime %w", err)
	}

	intensity := normalizeIntensity(req.Intensity)

	return models.ActivityRecord{
		UserID:    userID,
		Date:      day,
		Title:     title,
		StartTime: start,
		EndTime:   end,
		Intensity: intensity,
	}, nil
}

func buildActivityUpdates(req activityRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{}

	if strings.TrimSpace(req.Date) != "" {
		day, err := parseDate(req.Date)
		if err != nil {
			return nil, err
		}
		updates["date"] = day
	}
	if strings.TrimSpace(req.Title) != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}

	if strings.TrimSpace(req.StartTime) != "" || strings.TrimSpace(req.StartTimeAlt) != "" {
		start := strings.TrimSpace(firstNotEmpty(req.StartTime, req.StartTimeAlt))
		if err := validateClock(start); err != nil {
			return nil, fmt.Errorf("startTime %w", err)
		}
		updates["start_time"] = start
	}

	if strings.TrimSpace(req.EndTime) != "" || strings.TrimSpace(req.EndTimeAlt) != "" {
		end := strings.TrimSpace(firstNotEmpty(req.EndTime, req.EndTimeAlt))
		if err := validateClock(end); err != nil {
			return nil, fmt.Errorf("endTime %w", err)
		}
		updates["end_time"] = end
	}

	if strings.TrimSpace(req.Intensity) != "" {
		updates["intensity"] = normalizeIntensity(req.Intensity)
	}

	return updates, nil
}

const dateLayout = "2006-01-02"

func parseDate(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, errors.New("date is required in YYYY-MM-DD format.")
	}
	date, err := time.Parse(dateLayout, text)
	if err != nil {
		return time.Time{}, errors.New("Invalid date format. Expected YYYY-MM-DD.")
	}
	return date, nil
}

func normalizeMealType(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "breakfast", "lunch", "dinner", "snack":
		return strings.ToLower(strings.TrimSpace(raw)), nil
	default:
		return "", errors.New("meal type must be one of breakfast/lunch/dinner/snack.")
	}
}

func normalizeIntensity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func validateClock(raw string) error {
	if len(raw) != 5 {
		return errors.New("invalid time format. Expected HH:MM.")
	}
	if _, err := time.Parse("15:04", raw); err != nil {
		return errors.New("invalid time format. Expected HH:MM.")
	}
	return nil
}

func parseUintParam(raw string, fieldName string) (uint, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}

	id64, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid", fieldName)
	}
	return uint(id64), nil
}

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	userID, ok := userIDVal.(uint)
	return userID, ok
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalizedKey := strings.ToLower(trimmed)
		if _, exists := seen[normalizedKey]; exists {
			continue
		}
		seen[normalizedKey] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func ensureNonNilSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	return items
}

func weekRange(day time.Time) (time.Time, time.Time) {
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := day.AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 6)
	return start, end
}

func firstNotEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
