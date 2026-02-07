# Database Setup Guide

## Prerequisites
- Docker and Docker Compose installed
- Go 1.24.4 or later

## Step 1: Configure PostgreSQL Environment

Create a `.env.postgres` file in the root directory (copy from `env.postgres.example`):

```bash
cp env.postgres.example .env.postgres
```

Edit `.env.postgres` with your desired credentials:
```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=lazeez
```

## Step 2: Start PostgreSQL with Docker Compose

```bash
docker-compose up -d
```

This will start PostgreSQL on port `5432`.

## Step 3: Configure Application Environment

Create a `.env` file in the root directory (copy from `env.example`):

```bash
cp env.example .env
```

Edit `.env` with your database connection details:

**Option 1: Using DATABASE_URL (Recommended)**
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/lazeez?sslmode=disable
```

**Option 2: Using Individual Variables**
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=lazeez
```

Also set:
```
JWT_SECRET_KEY=your-secret-key-here
PORT=8080
```

## Step 4: Verify Database Connection

Check if PostgreSQL is running:
```bash
docker-compose ps
```

Test the connection:
```bash
docker-compose exec db psql -U postgres -d lazeez
```

## Step 5: Run the Application

```bash
go run cmd/main.go
```

The application will automatically connect to the PostgreSQL database using the environment variables.

## Troubleshooting

### Connection Refused
- Make sure Docker Compose is running: `docker-compose ps`
- Check if port 5432 is available: `lsof -i :5432`
- Verify `.env.postgres` credentials match your `.env` file

### Database Not Found
- The database is created automatically when PostgreSQL starts
- If needed, create it manually: `docker-compose exec db createdb -U postgres lazeez`

### Environment Variables Not Loading
- Make sure `.env` file exists in the root directory
- Check that variable names match exactly (case-sensitive)
- Restart the application after changing `.env` file
