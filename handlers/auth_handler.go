package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"meal_back/models"
	"meal_back/pkg/auth"
	"meal_back/pkg/response"
	"meal_back/stores"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	codeInvalidParam     = 10001
	codeRegisterFailed   = 10002
	codeUsernameExists   = 10003
	codeLoginFailed      = 10004
	codeAuthFailed       = 10005
	codeTokenIssueFailed = 10006
	codeUserNotFound     = 10007
)

// AuthHandler 负责认证相关接口（注册、登录、刷新、退出、用户信息）。
type AuthHandler struct {
	db             *gorm.DB
	tokenService   *auth.TokenService
	tokenBlacklist *stores.TokenBlacklistStore
}

// NewAuthHandler 创建 AuthHandler。
func NewAuthHandler(db *gorm.DB, jwtSecret string, tokenBlacklist *stores.TokenBlacklistStore) *AuthHandler {
	return &AuthHandler{
		db:             db,
		tokenService:   auth.NewTokenService(jwtSecret),
		tokenBlacklist: tokenBlacklist,
	}
}

// registerRequest 注册请求体。
// 说明：username 是必须字段，email 为可选字段。
type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// loginRequest 登录请求体。
// 当前登录方式：用户名 + 密码。
type loginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// upsertProfileRequest 兼容前端 snake_case / camelCase / legacy 字段。
type upsertProfileRequest struct {
	HeightCM  *float64 `json:"height_cm"`
	HeightCm  *float64 `json:"heightCm"`
	WeightKG  *float64 `json:"weight_kg"`
	WeightKg  *float64 `json:"weightKg"`
	Goal      string   `json:"fitness_goal"`
	GoalCamel string   `json:"fitnessGoal"`

	TrainingExperience      []string `json:"training_experience"`
	TrainingExperienceCamel []string `json:"trainingExperience"`
	ExerciseExperience      string   `json:"exercise_experience"`

	MonthlyFoodBudget      *int64 `json:"monthly_food_budget"`
	MonthlyFoodBudgetCamel *int64 `json:"monthlyFoodBudget"`
	MonthlyDietBudget      *int64 `json:"monthly_diet_budget"`
}

// Register 处理用户注册。
// 路由：POST /api/v1/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "Invalid request payload. Please check username/email/password.")
		return
	}

	username := normalizeUsername(req.Username)
	if username == "" {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "Username cannot be empty.")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeRegisterFailed, "Failed to hash password.")
		return
	}

	var emailPtr *string
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email != "" {
		emailPtr = &email
	}

	user := models.User{
		Username:     username,
		Email:        emailPtr,
		PasswordHash: string(passwordHash),
	}

	// 使用事务保证“创建用户 + 初始化资料”原子性。
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		profile := models.UserProfile{UserID: user.ID}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if isDuplicateKeyError(err) {
			if isUsernameDuplicate(err) {
				response.Error(c, http.StatusBadRequest, codeUsernameExists, "Username already exists.")
				return
			}
			response.Error(c, http.StatusBadRequest, codeRegisterFailed, "Email already exists.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeRegisterFailed, "Registration failed.")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// Login 处理用户登录。
// 路由：POST /api/v1/login
// 行为：
// 1. 校验用户名密码。
// 2. 签发 access/refresh token。
// 3. 将登录会话写入 user_sessions 表（持久化登录状态）。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "Invalid request payload. Please check username/password.")
		return
	}

	username := normalizeUsername(req.Username)
	var user models.User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusUnauthorized, codeLoginFailed, "Invalid username or password.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeLoginFailed, "Login failed.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, codeLoginFailed, "Invalid username or password.")
		return
	}

	accessToken, accessClaims, err := h.tokenService.CreateAccessToken(user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeTokenIssueFailed, "Failed to issue access token.")
		return
	}
	refreshToken, refreshClaims, err := h.tokenService.CreateRefreshToken(user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeTokenIssueFailed, "Failed to issue refresh token.")
		return
	}

	sessionID, err := randomID(16)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeTokenIssueFailed, "Failed to create session.")
		return
	}

	session := models.UserSession{
		UserID:           user.ID,
		SessionID:        sessionID,
		AccessTokenJTI:   accessClaims.ID,
		RefreshTokenJTI:  refreshClaims.ID,
		AccessExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshExpiresAt: refreshClaims.ExpiresAt.Time,
		LastSeenAt:       time.Now(),
		ClientIP:         c.ClientIP(),
		UserAgent:        truncateString(c.Request.UserAgent(), 512),
	}
	if err := h.db.Create(&session).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeLoginFailed, "Failed to persist login session.")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token":         accessToken,
		"session_id":    sessionID,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// RefreshToken 使用 refresh token 换取新的 access token。
// 路由：POST /api/v1/refresh
// 行为：
// 1. 验证 refresh token。
// 2. 根据 refresh_token_jti 找到持久化会话并检查未撤销。
// 3. 轮换 access token jti 并更新会话记录。
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "refresh_token is required.")
		return
	}

	claims, err := h.tokenService.ParseToken(strings.TrimSpace(req.RefreshToken))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Invalid or expired refresh token.")
		return
	}
	if claims.TokenType != auth.TokenTypeRefresh {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Invalid token type.")
		return
	}
	if h.tokenBlacklist != nil && h.tokenBlacklist.IsBlacklisted(claims.ID) {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Refresh token is revoked.")
		return
	}

	var session models.UserSession
	if err := h.db.Where("refresh_token_jti = ? AND is_revoked = ?", claims.ID, false).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Session not found or revoked. Please log in again.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeAuthFailed, "Failed to validate refresh session.")
		return
	}

	accessToken, accessClaims, err := h.tokenService.CreateAccessToken(claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, codeTokenIssueFailed, "Failed to issue access token.")
		return
	}

	if err := h.db.Model(&session).Updates(map[string]interface{}{
		"access_token_jti":  accessClaims.ID,
		"access_expires_at": accessClaims.ExpiresAt.Time,
		"last_seen_at":      time.Now(),
	}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeAuthFailed, "Failed to update session.")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"access_token": accessToken, "token": accessToken})
}

// Logout 让当前 access token 对应会话失效。
// 路由：POST /api/v1/private/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenJTIVal, ok := c.Get("tokenJTI")
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing token context.")
		return
	}
	expVal, ok := c.Get("tokenExpiresAt")
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing token expiration context.")
		return
	}

	tokenJTI, _ := tokenJTIVal.(string)
	expiresAt, _ := expVal.(int64)
	if tokenJTI == "" || expiresAt <= 0 {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Invalid token context.")
		return
	}

	now := time.Now()
	if err := h.db.Model(&models.UserSession{}).
		Where("access_token_jti = ? AND is_revoked = ?", tokenJTI, false).
		Updates(map[string]interface{}{"is_revoked": true, "revoked_at": now, "last_seen_at": now}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeAuthFailed, "Logout failed.")
		return
	}

	if h.tokenBlacklist != nil {
		h.tokenBlacklist.Add(tokenJTI, time.Unix(expiresAt, 0))
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Logged out."})
}

// Me 返回当前登录用户信息。
// 路由：GET /api/v1/private/me
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}

	userID, _ := userIDVal.(uint)
	var user models.User
	if err := h.db.Select("id", "username", "email", "created_at", "updated_at").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, codeUserNotFound, "User not found.")
			return
		}
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to query user.")
		return
	}

	var profile models.UserProfile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to query user profile.")
		return
	}

	var profilePtr *models.UserProfile
	if profile.ID != 0 {
		profilePtr = &profile
	}

	response.Success(c, http.StatusOK, buildUserPayload(user, profilePtr))
}

// UpsertProfile 更新当前登录用户资料，支持注册后问卷与后续修改。
// 路由：
// - PUT  /api/v1/users/me/profile
// - POST /api/v1/users/me/profile
// - PUT  /api/v1/private/me/profile
// - POST /api/v1/private/me/profile
func (h *AuthHandler) UpsertProfile(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		response.Error(c, http.StatusUnauthorized, codeAuthFailed, "Missing user identity in context.")
		return
	}
	userID, _ := userIDVal.(uint)

	var req upsertProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "Invalid request payload. Please check profile fields.")
		return
	}

	heightProvided := req.HeightCM != nil || req.HeightCm != nil
	weightProvided := req.WeightKG != nil || req.WeightKg != nil
	trainingProvided := req.TrainingExperience != nil || req.TrainingExperienceCamel != nil || strings.TrimSpace(req.ExerciseExperience) != ""
	goalRaw := strings.TrimSpace(req.Goal)
	if goalRaw == "" {
		goalRaw = strings.TrimSpace(req.GoalCamel)
	}
	goalProvided := goalRaw != ""
	budgetProvided := req.MonthlyFoodBudget != nil || req.MonthlyFoodBudgetCamel != nil || req.MonthlyDietBudget != nil

	if !heightProvided && !weightProvided && !trainingProvided && !goalProvided && !budgetProvided {
		response.Error(c, http.StatusBadRequest, codeInvalidParam, "Provide at least one updatable field.")
		return
	}

	updates := map[string]interface{}{}

	if heightProvided {
		height := firstNonNilFloat64(req.HeightCM, req.HeightCm)
		if height == nil || *height < 80 || *height > 260 {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, "height_cm must be between 80 and 260.")
			return
		}
		updates["height_cm"] = *height
	}

	if weightProvided {
		weight := firstNonNilFloat64(req.WeightKG, req.WeightKg)
		if weight == nil || *weight < 20 || *weight > 300 {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, "weight_kg must be between 20 and 300.")
			return
		}
		updates["weight_kg"] = *weight
	}

	if trainingProvided {
		rawTrainings := req.TrainingExperience
		if len(rawTrainings) == 0 {
			rawTrainings = req.TrainingExperienceCamel
		}
		if len(rawTrainings) == 0 && strings.TrimSpace(req.ExerciseExperience) != "" {
			rawTrainings = splitTrainingInput(req.ExerciseExperience)
		}

		trainings, err := normalizeTrainingExperience(rawTrainings)
		if err != nil {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, err.Error())
			return
		}
		if len(trainings) == 0 {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, "At least one training experience is required.")
			return
		}
		updates["exercise_experience"] = strings.Join(trainings, ",")
	}

	if goalProvided {
		goal, ok := normalizeFitnessGoal(goalRaw)
		if !ok {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, "fitness_goal must be one of lose_weight/build_muscle/maintain_shape.")
			return
		}
		updates["fitness_goal"] = goal
	}

	if budgetProvided {
		budget := firstNonNilInt64(req.MonthlyFoodBudget, req.MonthlyFoodBudgetCamel, req.MonthlyDietBudget)
		if budget == nil || *budget < 100 {
			response.Error(c, http.StatusBadRequest, codeInvalidParam, "monthly_food_budget must be at least 100.")
			return
		}
		updates["monthly_diet_budget"] = *budget
	}

	if err := h.db.Where("user_id = ?", userID).FirstOrCreate(&models.UserProfile{UserID: userID}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to initialize user profile.")
		return
	}

	if err := h.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to save user profile.")
		return
	}

	var user models.User
	if err := h.db.Select("id", "username", "email", "created_at", "updated_at").First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to query user.")
		return
	}

	var profile models.UserProfile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, codeUserNotFound, "Failed to query user profile.")
		return
	}

	response.Success(c, http.StatusOK, buildUserPayload(user, &profile))
}

// isDuplicateKeyError 判断是否为唯一索引冲突。
func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

func isUsernameDuplicate(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "users_username") || strings.Contains(msg, "username")
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func randomID(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func truncateString(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}
	return input[:maxLen]
}

func buildUserPayload(user models.User, profile *models.UserProfile) gin.H {
	payload := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"profile":    gin.H{},
	}

	if profile == nil {
		return payload
	}

	trainingExperience := parseTrainingExperienceText(profile.ExerciseExperience)

	profileData := gin.H{
		"nickname":            profile.Nickname,
		"avatar":              profile.Avatar,
		"bio":                 profile.Bio,
		"height_cm":           profile.HeightCM,
		"weight_kg":           profile.WeightKG,
		"allergies":           ensureNonNilSlice(profile.Allergies),
		"dietary_preferences": ensureNonNilSlice(profile.DietaryPreferences),
		"exercise_experience": profile.ExerciseExperience,
		"training_experience": trainingExperience,
		"fitness_goal":        profile.FitnessGoal,
		"monthly_diet_budget": profile.MonthlyDietBudget,
		"monthly_food_budget": profile.MonthlyDietBudget,
	}

	payload["height_cm"] = profile.HeightCM
	payload["weight_kg"] = profile.WeightKG
	payload["allergies"] = ensureNonNilSlice(profile.Allergies)
	payload["dietary_preferences"] = ensureNonNilSlice(profile.DietaryPreferences)
	payload["exercise_experience"] = profile.ExerciseExperience
	payload["training_experience"] = trainingExperience
	payload["fitness_goal"] = profile.FitnessGoal
	payload["monthly_diet_budget"] = profile.MonthlyDietBudget
	payload["monthly_food_budget"] = profile.MonthlyDietBudget
	payload["profile"] = profileData

	return payload
}

func parseTrainingExperienceText(raw string) []string {
	items := splitTrainingInput(raw)
	if len(items) == 0 {
		return []string{}
	}

	normalized, err := normalizeTrainingExperience(items)
	if err != nil {
		return []string{}
	}
	return normalized
}

func splitTrainingInput(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', '、', ' ', '\t', '\n', '\r':
			return true
		default:
			return false
		}
	})
}

func normalizeTrainingExperience(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}

	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, item := range values {
		parts := splitTrainingInput(item)
		if len(parts) == 0 {
			parts = []string{item}
		}
		for _, part := range parts {
			normalized, ok := normalizeTrainingToken(part)
			if !ok {
				return nil, errors.New("training_experience must be one of fitness/yoga/pilates/climbing.")
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}

	return out, nil
}

func normalizeTrainingToken(raw string) (string, bool) {
	token := strings.ToLower(strings.TrimSpace(raw))
	switch token {
	case "fitness":
		return "fitness", true
	case "yoga":
		return "yoga", true
	case "pilates":
		return "pilates", true
	case "climbing":
		return "climbing", true
	default:
		return "", false
	}
}

func normalizeFitnessGoal(raw string) (string, bool) {
	token := strings.ToLower(strings.TrimSpace(raw))
	switch token {
	case "lose_weight":
		return "lose_weight", true
	case "build_muscle":
		return "build_muscle", true
	case "maintain_shape":
		return "maintain_shape", true
	default:
		return "", false
	}
}

func firstNonNilFloat64(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstNonNilInt64(values ...*int64) *int64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
