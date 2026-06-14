# TaskFlow

Collaborative task management system built with Go.

## Features

- User Authentication (JWT)
- Team Management (Create, Add Members)
- Task Management (CRUD, Assignment, Status tracking)
- RBAC (Owner, Manager, Member roles)
- Task History & Audit Log
- Analytics
- Redis Caching
- Simple Web Frontend

## Tech Stack

- **Go 1.24**
- **Gin** (HTTP framework)
- **PostgreSQL** (Database)
- **Redis** (Cache)
- **Docker** & **Docker Compose**
- **Clean Architecture**

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.24 (for local development)

### Running the Project

1. Clone the repository
2. Run with Docker Compose:
   ```bash
   make docker-up
   ```
3. Open `http://localhost:8080` in your browser.

### Development

- Run tests: `make test`
- Run locally: `make run` (requires local Postgres and Redis)
- Run migrations: `make migrate-up`

## Project Structure

- `cmd/server`: Application entry point
- `internal/domain`: Core entities
- `internal/repository`: Data access layer (Postgres)
- `internal/service`: Business logic
- `internal/transport/http`: API handlers and routes
- `internal/cache`: Redis cache layer
- `migrations`: SQL schema migrations
- `web/static`: Simple frontend assets
