# meal_back

Run `meal_back` with Docker Compose only.

## Prerequisites

- Docker Engine / Docker Desktop
- Docker Compose v2 (`docker compose`)

## Quick start

```bash
make setup
make run-env
```

API base URL:

```text
http://localhost:8080/api/v1
```

## Common commands

```bash
make up        # start in background
make logs      # stream logs
make ps        # check status
make down      # stop containers, keep db data
make teardown  # stop containers and remove db volume
```
