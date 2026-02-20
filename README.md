# lazeez-core

## Local development

Run the app locally with the database in Docker:

```bash
# 1. Start the database
make dev-db
# or: docker compose up -d db

# 2. Run the app (from project root so .env is loaded)
make dev
# or: go run ./cmd
```

Ensure `.env` exists (copy from `.env.example` if needed). The app connects to `localhost:5432` when running outside Docker.