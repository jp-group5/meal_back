package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"meal_back/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type userTarget struct {
	ID       uint
	Username string
}

type response struct {
	OK      bool        `json:"ok"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type settableString struct {
	Value string
	IsSet bool
}

func (s *settableString) String() string {
	return s.Value
}

func (s *settableString) Set(value string) error {
	s.Value = strings.TrimSpace(value)
	s.IsSet = true
	return nil
}

type settableFloat64 struct {
	Value float64
	IsSet bool
}

func (s *settableFloat64) String() string {
	if !s.IsSet {
		return ""
	}
	return strconv.FormatFloat(s.Value, 'f', -1, 64)
}

func (s *settableFloat64) Set(value string) error {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return fmt.Errorf("无效浮点数: %w", err)
	}
	s.Value = parsed
	s.IsSet = true
	return nil
}

type settableInt64 struct {
	Value int64
	IsSet bool
}

func (s *settableInt64) String() string {
	if !s.IsSet {
		return ""
	}
	return strconv.FormatInt(s.Value, 10)
}

func (s *settableInt64) Set(value string) error {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return fmt.Errorf("无效整数: %w", err)
	}
	s.Value = parsed
	s.IsSet = true
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		exitWithError("请提供子命令")
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		printUsage()
		return
	}

	db, err := connectDB()
	if err != nil {
		exitWithError(err.Error())
	}

	switch cmd {
	case "list":
		err = cmdList(db, args)
	case "show":
		err = cmdShow(db, args)
	case "create":
		err = cmdCreate(db, args)
	case "update":
		err = cmdUpdate(db, args)
	case "set-password":
		err = cmdSetPassword(db, args)
	case "upsert-profile":
		err = cmdUpsertProfile(db, args)
	case "delete":
		err = cmdDelete(db, args)
	default:
		printUsage()
		exitWithError("未知子命令: " + cmd)
	}

	if err != nil {
		exitWithError(err.Error())
	}
}

func connectDB() (*gorm.DB, error) {
	dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
	if dsn == "" {
		return nil, errors.New("缺少环境变量 DB_DSN")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}

func cmdList(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	limit := fs.Int("limit", 20, "每页条数")
	offset := fs.Int("offset", 0, "偏移量")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *limit <= 0 || *limit > 200 {
		return errors.New("limit 取值范围为 1~200")
	}
	if *offset < 0 {
		return errors.New("offset 不能小于 0")
	}

	type row struct {
		ID        uint    `json:"id"`
		Username  string  `json:"username"`
		Email     *string `json:"email,omitempty"`
		CreatedAt string  `json:"created_at"`
		UpdatedAt string  `json:"updated_at"`
	}

	var users []models.User
	if err := db.Select("id", "username", "email", "created_at", "updated_at").
		Order("id ASC").
		Limit(*limit).
		Offset(*offset).
		Find(&users).Error; err != nil {
		return fmt.Errorf("查询用户列表失败: %w", err)
	}

	out := make([]row, 0, len(users))
	for _, u := range users {
		out = append(out, row{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	printJSON(response{OK: true, Data: out})
	return nil
}

func cmdShow(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.Uint("id", 0, "用户ID")
	username := fs.String("username", "", "用户名")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := parseTarget(*id, *username)
	if err != nil {
		return err
	}

	user, profile, err := loadUserAndProfile(db, target)
	if err != nil {
		return err
	}

	printJSON(response{OK: true, Data: map[string]interface{}{
		"user":    user,
		"profile": profile,
	}})
	return nil
}

func cmdCreate(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	username := fs.String("username", "", "用户名")
	email := fs.String("email", "", "邮箱（可选）")
	password := fs.String("password", "", "密码")
	if err := fs.Parse(args); err != nil {
		return err
	}

	normalizedUsername := normalizeUsername(*username)
	if normalizedUsername == "" {
		return errors.New("username 不能为空")
	}
	if len(*password) < 8 || len(*password) > 72 {
		return errors.New("password 长度需在 8~72 之间")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	var emailPtr *string
	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))
	if normalizedEmail != "" {
		emailPtr = &normalizedEmail
	}

	user := models.User{
		Username:     normalizedUsername,
		Email:        emailPtr,
		PasswordHash: string(hashed),
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		profile := models.UserProfile{UserID: user.ID}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	printJSON(response{OK: true, Message: "创建成功", Data: map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	}})
	return nil
}

func cmdUpdate(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.Uint("id", 0, "用户ID")
	username := fs.String("username", "", "用户名")
	newUsername := fs.String("new-username", "", "新用户名")
	email := fs.String("email", "", "新邮箱")
	clearEmail := fs.Bool("clear-email", false, "清空邮箱")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := parseTarget(*id, *username)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if trimmed := normalizeUsername(*newUsername); trimmed != "" {
		updates["username"] = trimmed
	}

	if *clearEmail {
		updates["email"] = nil
	} else if trimmed := strings.ToLower(strings.TrimSpace(*email)); trimmed != "" {
		updates["email"] = trimmed
	}

	if len(updates) == 0 {
		return errors.New("没有可更新字段，请至少提供 --new-username / --email / --clear-email")
	}

	user, err := findUser(db, target)
	if err != nil {
		return err
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	if err := db.Select("id", "username", "email", "updated_at").First(&user, user.ID).Error; err != nil {
		return fmt.Errorf("读取更新后用户失败: %w", err)
	}

	printJSON(response{OK: true, Message: "更新成功", Data: user})
	return nil
}

func cmdSetPassword(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("set-password", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.Uint("id", 0, "用户ID")
	username := fs.String("username", "", "用户名")
	password := fs.String("password", "", "新密码")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := parseTarget(*id, *username)
	if err != nil {
		return err
	}
	if len(*password) < 8 || len(*password) > 72 {
		return errors.New("password 长度需在 8~72 之间")
	}

	user, err := findUser(db, target)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	if err := db.Model(&user).Update("password_hash", string(hashed)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	printJSON(response{OK: true, Message: "密码更新成功", Data: map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
	}})
	return nil
}

func cmdUpsertProfile(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("upsert-profile", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.Uint("id", 0, "用户ID")
	username := fs.String("username", "", "用户名")

	var heightCM settableFloat64
	var weightKG settableFloat64
	var monthlyBudget settableInt64
	var exercise settableString
	var fitnessGoal settableString
	var nickname settableString
	var avatar settableString
	var bio settableString

	fs.Var(&heightCM, "height-cm", "身高（厘米）")
	fs.Var(&weightKG, "weight-kg", "体重（公斤）")
	fs.Var(&exercise, "exercise", "运动经验，如: 健身,瑜伽,普拉提")
	fs.Var(&fitnessGoal, "goal", "目标：减重/增肌/维持身材")
	fs.Var(&monthlyBudget, "budget", "每月饮食费上限")
	fs.Var(&nickname, "nickname", "昵称")
	fs.Var(&avatar, "avatar", "头像 URL")
	fs.Var(&bio, "bio", "简介")

	clearHeight := fs.Bool("clear-height", false, "清空身高")
	clearWeight := fs.Bool("clear-weight", false, "清空体重")
	clearBudget := fs.Bool("clear-budget", false, "清空月饮食预算")
	clearExercise := fs.Bool("clear-exercise", false, "清空运动经验")
	clearGoal := fs.Bool("clear-goal", false, "清空健身目标")
	clearNickname := fs.Bool("clear-nickname", false, "清空昵称")
	clearAvatar := fs.Bool("clear-avatar", false, "清空头像")
	clearBio := fs.Bool("clear-bio", false, "清空简介")

	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := parseTarget(*id, *username)
	if err != nil {
		return err
	}

	user, err := findUser(db, target)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if heightCM.IsSet {
		updates["height_cm"] = heightCM.Value
	}
	if *clearHeight {
		updates["height_cm"] = nil
	}
	if weightKG.IsSet {
		updates["weight_kg"] = weightKG.Value
	}
	if *clearWeight {
		updates["weight_kg"] = nil
	}
	if monthlyBudget.IsSet {
		updates["monthly_diet_budget"] = monthlyBudget.Value
	}
	if *clearBudget {
		updates["monthly_diet_budget"] = nil
	}
	if exercise.IsSet {
		updates["exercise_experience"] = exercise.Value
	}
	if *clearExercise {
		updates["exercise_experience"] = ""
	}
	if fitnessGoal.IsSet {
		updates["fitness_goal"] = fitnessGoal.Value
	}
	if *clearGoal {
		updates["fitness_goal"] = ""
	}
	if nickname.IsSet {
		updates["nickname"] = nickname.Value
	}
	if *clearNickname {
		updates["nickname"] = ""
	}
	if avatar.IsSet {
		updates["avatar"] = avatar.Value
	}
	if *clearAvatar {
		updates["avatar"] = ""
	}
	if bio.IsSet {
		updates["bio"] = bio.Value
	}
	if *clearBio {
		updates["bio"] = ""
	}

	if len(updates) == 0 {
		return errors.New("没有可更新字段")
	}

	if err := db.Where("user_id = ?", user.ID).FirstOrCreate(&models.UserProfile{UserID: user.ID}).Error; err != nil {
		return fmt.Errorf("初始化 profile 失败: %w", err)
	}

	if err := db.Model(&models.UserProfile{}).Where("user_id = ?", user.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新 profile 失败: %w", err)
	}

	var profile models.UserProfile
	if err := db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		return fmt.Errorf("读取 profile 失败: %w", err)
	}

	printJSON(response{OK: true, Message: "资料更新成功", Data: profile})
	return nil
}

func cmdDelete(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.Uint("id", 0, "用户ID")
	username := fs.String("username", "", "用户名")
	hard := fs.Bool("hard", false, "硬删除（Unscoped）")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := parseTarget(*id, *username)
	if err != nil {
		return err
	}

	user, err := findUser(db, target)
	if err != nil {
		return err
	}

	if *hard {
		if err := db.Unscoped().Where("user_id = ?", user.ID).Delete(&models.UserProfile{}).Error; err != nil {
			return fmt.Errorf("硬删除 profile 失败: %w", err)
		}
		if err := db.Unscoped().Where("user_id = ?", user.ID).Delete(&models.UserSession{}).Error; err != nil {
			return fmt.Errorf("硬删除 session 失败: %w", err)
		}
		if err := db.Unscoped().Delete(&user).Error; err != nil {
			return fmt.Errorf("硬删除用户失败: %w", err)
		}
		printJSON(response{OK: true, Message: "硬删除成功"})
		return nil
	}

	if err := db.Where("user_id = ?", user.ID).Delete(&models.UserProfile{}).Error; err != nil {
		return fmt.Errorf("删除 profile 失败: %w", err)
	}
	if err := db.Where("user_id = ?", user.ID).Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("删除 session 失败: %w", err)
	}
	if err := db.Delete(&user).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	printJSON(response{OK: true, Message: "软删除成功"})
	return nil
}

func parseTarget(id uint, username string) (userTarget, error) {
	trimmedUsername := normalizeUsername(username)
	if id == 0 && trimmedUsername == "" {
		return userTarget{}, errors.New("请提供 --id 或 --username")
	}
	if id != 0 && trimmedUsername != "" {
		return userTarget{}, errors.New("--id 和 --username 只能二选一")
	}
	return userTarget{ID: id, Username: trimmedUsername}, nil
}

func findUser(db *gorm.DB, target userTarget) (models.User, error) {
	var user models.User
	var err error
	if target.ID != 0 {
		err = db.First(&user, target.ID).Error
	} else {
		err = db.Where("username = ?", target.Username).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("用户不存在")
		}
		return models.User{}, fmt.Errorf("查询用户失败: %w", err)
	}
	return user, nil
}

func loadUserAndProfile(db *gorm.DB, target userTarget) (models.User, *models.UserProfile, error) {
	user, err := findUser(db, target)
	if err != nil {
		return models.User{}, nil, err
	}

	var profile models.UserProfile
	if err := db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, nil, nil
		}
		return models.User{}, nil, fmt.Errorf("查询用户资料失败: %w", err)
	}
	return user, &profile, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 编码失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}

func printUsage() {
	usage := `userctl: 数据库用户账号管理工具

用法:
  go run ./cmd/userctl <command> [flags]

命令:
  list             列出用户
  show             查看单个用户 + profile
  create           创建用户并初始化 profile
  update           更新 username/email
  set-password     重置密码
  upsert-profile   更新或创建 profile（含身高/体重/运动经验/目标/预算）
  delete           删除用户（默认软删除，可选 --hard）

示例:
  go run ./cmd/userctl list --limit 50
  go run ./cmd/userctl show --username testuser
  go run ./cmd/userctl create --username alice --password 12345678 --email a@demo.com
  go run ./cmd/userctl update --id 1 --new-username alice2 --email alice2@demo.com
  go run ./cmd/userctl set-password --username alice2 --password newpassword123
  go run ./cmd/userctl upsert-profile --username alice2 --height-cm 172 --weight-kg 65 --exercise "健身,瑜伽" --goal 减重 --budget 2000
  go run ./cmd/userctl delete --id 1
`
	fmt.Println(usage)
}

func exitWithError(msg string) {
	printJSON(response{OK: false, Message: msg})
	os.Exit(1)
}
