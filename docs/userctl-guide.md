# userctl 操作手册（用户账号增删改查）

`userctl` 是后端仓库内置的数据库用户管理工具，支持：

- 用户增删改查
- 密码重置
- 用户问卷资料（身高/体重/运动经验/目标/预算）更新

## 1. 前置条件

先设置数据库连接：

```bash
export DB_DSN='host=127.0.0.1 port=5432 user=teiyou dbname=meal_back sslmode=disable TimeZone=Asia/Tokyo'
```

命令入口有两种：

```bash
# 推荐（通过 Makefile）
make userctl ARGS='<命令>'

# 直接运行
go run ./cmd/userctl <命令>
```

---

## 2. 查（Read）

### 2.1 列出用户

```bash
make userctl ARGS='list --limit 50 --offset 0'
```

参数说明：

- `--limit`：每页条数（1~200）
- `--offset`：偏移量

### 2.2 查看单个用户

```bash
make userctl ARGS='show --username alice'
make userctl ARGS='show --id 1'
```

---

## 3. 增（Create）

创建账号（同时初始化 profile）：

```bash
make userctl ARGS='create --username alice --password 12345678 --email alice@example.com'
```

参数说明：

- `--username`：必填
- `--password`：必填，8~72 位
- `--email`：可选

---

## 4. 改（Update）

### 4.1 修改用户名/邮箱

```bash
make userctl ARGS='update --username alice --new-username alice2 --email alice2@example.com'
```

### 4.2 清空邮箱

```bash
make userctl ARGS='update --id 1 --clear-email'
```

### 4.3 重置密码

```bash
make userctl ARGS='set-password --username alice2 --password NewPass123'
```

注意：数据库只存 `password_hash`，不能查看明文密码。

### 4.4 更新问卷资料（可反复修改）

```bash
make userctl ARGS='upsert-profile --username alice2 --height-cm 172 --weight-kg 65 --exercise "健身,瑜伽" --goal 减重 --budget 2500'
```

支持字段：

- `--height-cm`：身高
- `--weight-kg`：体重
- `--exercise`：运动经验（逗号分隔文本）
- `--goal`：目标（如 `减重` / `增肌` / `维持身材`）
- `--budget`：每月饮食费上限
- `--nickname` / `--avatar` / `--bio`：其他资料

### 4.5 清空问卷字段

```bash
make userctl ARGS='upsert-profile --username alice2 --clear-height --clear-weight --clear-budget --clear-exercise --clear-goal'
```

也可按需清空：

- `--clear-height`
- `--clear-weight`
- `--clear-budget`
- `--clear-exercise`
- `--clear-goal`
- `--clear-nickname`
- `--clear-avatar`
- `--clear-bio`

---

## 5. 删（Delete）

### 5.1 软删除（默认）

```bash
make userctl ARGS='delete --username alice2'
```

### 5.2 硬删除（物理删除）

```bash
make userctl ARGS='delete --id 1 --hard'
```

---

## 6. 一键清理所有已注册用户（危险）

该脚本会直接清空：

- `users`
- `user_profiles`
- `user_sessions`

执行命令（必须显式确认）：

```bash
CONFIRM=YES make cleanup-users
```

或直接运行脚本：

```bash
DB_DSN='host=127.0.0.1 port=5432 user=teiyou dbname=meal_back sslmode=disable TimeZone=Asia/Tokyo' \
./scripts/cleanup_all_users.sh --yes
```

---

## 7. 参数规则

多数命令用于定位用户时，`--id` 和 `--username` 二选一，不能同时传。

---

## 8. 常见问题

### 8.1 `ERROR: DB_DSN 未设置`

先执行：

```bash
export DB_DSN='host=127.0.0.1 port=5432 user=teiyou dbname=meal_back sslmode=disable TimeZone=Asia/Tokyo'
```

### 8.2 `FATAL: role "postgres" does not exist`

本地数据库角色不是 `postgres`，把 `DB_DSN` 中 `user=postgres` 改为你本机实际角色（例如 `user=teiyou`）。

### 8.3 查看全部命令帮助

```bash
go run ./cmd/userctl help
```
