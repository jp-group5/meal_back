# Meal App 后端职责与实现规范（可复用）

## 1. 项目目标
- 用户在 Web 页面中结合日历记录两类信息：`行程(activity)` 与 `饮食(meal)`。
- 系统基于用户基础资料、最近饮食记录、本周行程，生成饮食建议。
- 后端职责：
1. 管理数据库中的用户与业务数据。
2. 提供 Go API 与前端交互。
3. 将“用户资料 + 饮食记录 + 行程”拼接为 AI 可用 prompt（JSON）。

## 2. 后端职责拆分
1. 账号与身份
- 用户注册、登录、刷新 token、登出、获取当前用户。
- JWT + 会话持久化（`user_sessions`）。

2. 用户资料
- 资料字段：身高、体重、运动经验、目标、月饮食预算。
- 偏好字段：过敏原（`allergies`）、饮食偏好（`dietary_preferences`）。

3. 饮食记录（Meals）
- 按日期查询、新增、更新、删除。
- 支持 AI 推荐落库字段（`source`、`recommendationId`）。

4. 行程记录（Activities）
- 按日期查询、新增、更新、删除。
- 记录强度（`low|medium|high`），用于推荐时的能量需求判断。

5. 推荐与 Prompt
- 输入日期，聚合：
- 用户资料（含目标、偏好、预算）
- 近 7 天饮食记录
- 本周行程（周一到周日）
- 输出：
- 前端可展示的推荐结果
- `prompt_json`（可直接传给 AI 模型）

## 3. 与前端对齐的 API（`/api/v1`）
- `PUT /users/me/preferences`
- `GET /meals?date=YYYY-MM-DD`
- `POST /meals`
- `PUT /meals/:id`
- `DELETE /meals/:id`
- `GET /activities?date=YYYY-MM-DD`
- `POST /activities`
- `PUT /activities/:id`
- `DELETE /activities/:id`
- `POST /recommendations`（body: `{ "date": "YYYY-MM-DD" }`）
- `GET /recommendations/prompt?date=YYYY-MM-DD`（调试 prompt）

说明：这些接口都要求 Bearer access token（与现有 `/private/*` 保持一致鉴权机制）。

## 4. 当前数据库核心表
- `users`
- `user_profiles`（新增 `allergies`, `dietary_preferences` JSONB）
- `user_sessions`
- `meal_records`
- `activity_records`

## 5. 推荐 Prompt JSON 结构（示例）
```json
{
  "metadata": {
    "generated_at": "2026-05-12T12:00:00+09:00",
    "target_date": "2026-05-12",
    "time_zone": "Asia/Tokyo"
  },
  "user": {
    "user_id": 1,
    "username": "alice",
    "height_cm": 168,
    "weight_kg": 56,
    "fitness_goal": "lose_weight",
    "training_experience": ["fitness", "yoga"],
    "monthly_food_budget": 2500,
    "allergies": ["peanut"],
    "dietary_preferences": ["high_protein"]
  },
  "context": {
    "recent_meals": [],
    "weekly_activities": [],
    "meal_stats": {
      "days_with_meals": 0,
      "total_meal_records": 0,
      "average_daily_calories": 0,
      "records_with_calories": 0,
      "calories_window_in_days": 7,
      "recent_meals_window_days": 7
    },
    "activity_intensity_stats": {
      "low": 0,
      "medium": 0,
      "high": 0,
      "unknown": 0
    }
  }
}
```

## 6. 以后可重复调用的任务模板（直接复制给 AI）
```text
你是我的 Go 后端协作助手。请在 meal_app 后端中完成以下目标：
1) 保持与前端 API 对齐（/meals, /activities, /recommendations, /users/me/preferences）。
2) 使用 PostgreSQL + GORM 管理用户、饮食、行程、偏好数据。
3) 推荐接口必须聚合“用户基础资料 + 最近7天饮食 + 本周行程”，并输出 prompt_json。
4) 所有新增接口需接入现有 JWT 鉴权与统一响应格式（code/message/data）。
5) 输出改动文件清单、关键字段说明、以及最小联调步骤。
```

## 7. 联调最小步骤
1. 配置环境变量：`DB_DSN`, `JWT_SECRET`。
2. 启动后端（默认 `:8080`）。
3. 前端 `VITE_API_BASE_URL` 指向 `http://localhost:8080/api/v1`。
4. 登录后调用：
- `POST /recommendations` 查看推荐结果与 `prompt_json`
- `GET /recommendations/prompt?date=...` 单独检查 prompt 组装内容

