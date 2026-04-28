# meal_back

Backend repository for `meal_app` (Go + Gin + GORM + PostgreSQL).

## Features

- JWT auth with persistent sessions (`register`, `login`, `refresh`, `logout`)
- User profile and preferences
- Meal records CRUD
- Activity records CRUD
- Recommendation context assembly (`prompt_json`) for AI calls

## API Base

`/api/v1`

## Environment Variables

- `DB_DSN`
- `JWT_SECRET`

## Run

```bash
go run .
```

Server default: `:8080`

## Documents

- Backend role/integration spec: [docs/backend-role-spec.md](./docs/backend-role-spec.md)
- User account CLI guide: [docs/userctl-guide.md](./docs/userctl-guide.md)
