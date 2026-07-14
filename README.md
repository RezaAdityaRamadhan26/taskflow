# TaskFlow - Advanced Kanban Task Management

A full-stack Kanban board application (Jira/Trello clone) built with modern technologies.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.26 + Fiber v2 |
| **Database** | Supabase (PostgreSQL) |
| **Schema & Migrations** | Prisma ORM (Node.js) |
| **Go DB Layer** | pgx v5 + sqlc |
| **Authentication** | JWT (Access + Refresh Token) |
| **Frontend** | Vite + React 19 + TypeScript + TailwindCSS |

## Project Structure

```
taskflow/
├── backend/          # Go Fiber API server
│   ├── cmd/server/   # Entry point
│   ├── internal/     # Application code (config, handler, service, repository, etc.)
│   ├── db/           # sqlc queries and config
│   └── tests/        # Integration tests
├── frontend/         # React + Vite SPA
│   └── src/          # Components, features, hooks, services, store, types, utils
└── prisma/           # Prisma schema & migrations
```

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Supabase account with a project

### Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your credentials
3. Run Prisma migrations: `cd prisma && npx prisma migrate dev`
4. Start the backend: `cd backend && go run cmd/server/main.go`
5. Start the frontend: `cd frontend && npm run dev`

## License

MIT
