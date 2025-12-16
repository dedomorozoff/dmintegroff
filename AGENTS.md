# AGENTS.md - Project Context for AI Agents

This file provides high-level context, architectural guidelines, and codebase conventions for AI agents working on **dmIntegroff**.

## Project Overview
**Name**: dmIntegroff
**Description**: A REST API integration management system built with Go. It allows users to configure connections between different APIs, transform data, and automatically transfer it via webhooks.
**Primary Language**: Go (Backend), HTML/JS/CSS (Frontend). 
**UI Language**: Russian (The interface and user-facing documentation are in Russian).

## Technology Stack
- **Backend**: Go 1.25+ (updated from 1.20)
- **Web Framework**: Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database**: GORM (`gorm.io/gorm`) supporting SQLite (default) and MySQL
- **Cache**: Redis support for webhook tests (optional, falls back to database)
- **Sessions**: Cookie-based sessions with `gin-contrib/sessions`
- **Logging**: Structured logging with `sirupsen/logrus`
- **Monitoring**: Prometheus metrics support (`prometheus/client_golang`)
- **Frontend**: Server-side rendered Go HTML Templates
- **Styling**: Vanilla CSS ("Modern CSS"), located in `static/css/modern.css`. No heavy CSS frameworks like Tailwind or Bootstrap are currently strictly enforced, mostly custom styles
- **Scripting**: Vanilla JavaScript, located in `static/js/` and inline in templates
- **Authentication**: Session-based with password hashing using `golang.org/x/crypto`
- **Testing**: `stretchr/testify` for unit tests, `DATA-DOG/go-sqlmock` for database mocking

## Architecture
The project follows a standard Go project layout:

- **`cmd/server/main.go`**: Application entry point
- **`internal/`**: Core application logic
    - **`controllers/`**: HTTP request handlers (MVC pattern)
    - **`models/`**: GORM database models (Integration, Project, User, etc.)
    - **`routes/`**: Route definitions and grouping
    - **`services/`**: Business logic encapsulation
    - **`database/`**: Database connection and initialization
    - **`logger/`**: Custom logging setup with structured logging
    - **`middleware/`**: Rate limiting, authentication, request logging
    - **`ai/`**: AI integration components (client, prompts, generators)
    - **`cache/`**: Redis caching layer
    - **`utils/`**: Helper functions (JSON parsing, template processing, etc.)
- **`templates/`**: HTML templates
    - **`pages/`**: Full page templates
    - **`partials/`**: Reusable template fragments (header, sidebar, footer)
- **`static/`**: Publicly accessible assets (CSS, JS, images)
- **`docs/`**: Project documentation (Markdown, primarily in Russian)
- **`tests/`**: Test files and test data
- **`logs/`**: Application logs directory
- **`db/`**: Database files (SQLite) and migrations
- **`.env`**: Configuration file (Environment variables)

## Development Guidelines

### Coding Style
- **Go**: Follow standard Go conventions (`gofmt`). Error handling should be explicit.
- **HTML/CSS**: Keep it semantic and accessible. Use the existing "modern" dark-themed aesthetic.
- **JavaScript**: Use modern ES6+ syntax. Avoid jQuery if vanilla JS suffices.

### Language & Localization
- **Code Comments**: English is preferred for code comments and commit messages, but Russian is acceptable if context dictates (e.g., explaining business logic specific to Russian integration nuances).
- **User Interface**: **MUST** be in **Russian**. All labels, buttons, and help text in `templates/` should be in Russian.
- **Documentation**: Project documentation in `docs/` is primarily in Russian.

### Key Features
- **Integrations**: Source API -> Webhook -> Transformation -> Target API
- **Multiple Outputs**: One webhook can have multiple outputs (mappings) to different Target APIs
- **Data Formats**: Support for JSON, XML, form-data, plain text input formats
- **Artificial Intelligence**: AI-powered integration creation and mapping generation (planned)
- **Webhooks**: Handling incoming data, validating multiple formats, syntax highlighting
- **Authentication**: Support for OAuth, Bearer tokens, Basic Auth for Target APIs
- **Testing**: Built-in webhook testing endpoint with request history
- **Analytics**: Dashboard with request statistics and system monitoring
- **Projects**: Grouping integrations into projects for better organization
- **Export/Import**: Configuration backup/restore and Git export functionality
- **Logging**: Comprehensive request/response logging, error tracking, and audit trails

## Common Tasks
- **Running the Server**: `go run cmd/server/main.go`
- **Running Tests**: `go test ./...` or `go test ./internal/...`
- **Database Migrations**: Handled automatically by GORM on startup (auto-migrate)
- **Configuration**: Managed via `.env` file (copy from `.env.example`)
- **Building**: `go build -o dmintegroff.exe cmd/server/main.go`
- **Logs**: Check `logs/app.log` or stdout for application logs
- **Redis**: Optional for webhook test caching, configure via `REDIS_URL` in `.env`

## Roadmap & Status
- **Current Focus**: AI-powered integration creation (revolutionary feature), request history, configuration validation
- **Recently Completed**: Multiple outputs per webhook, multi-format data support, webhook testing, export/import
- **Next Priority**: AI assistant for automatic integration creation using natural language
- **Future**: Docker support, GraphQL server mode, visual mapping editor
- **References**: See `docs/ROADMAP.md`, `docs/AI_SUMMARY.md`, and `docs/TODO_NEXT.md` for detailed plans

## Current Project Status

### ✅ Recently Implemented Features
- **Multiple Outputs**: One webhook can trigger multiple Target APIs with different mappings
- **Multi-Format Support**: JSON, XML, form-data, plain text input processing
- **Webhook Testing**: Built-in test endpoint with request capture and replay
- **Export/Import**: Configuration backup/restore and Git export functionality
- **Enhanced Analytics**: Dashboard with system stats and request monitoring
- **Project Organization**: Grouping integrations into projects

### 🔥 Critical Priority (Production Ready)
- **Request History**: Detailed logging and search of all webhook requests
- **Configuration Validation**: Test connections and validate settings before activation
- **Data Export**: CSV/JSON export for analytics and audit trails

### 🚀 Revolutionary Feature in Development
- **AI Assistant**: Natural language integration creation using OpenAI API
  - User describes task → AI creates complete integration automatically
  - Planned to reduce setup time from hours to minutes
  - See `docs/AI_INTEGRATION_PLAN.md` for detailed implementation plan

## Special Instructions for Agents
1. **Context Loading**: When starting a task, check `internal/routes/routes.go` to understand the URL structure and `internal/models/` for data structure
2. **UI Changes**: When modifying UI, ensure you are editing the correct template in `templates/` and that styles in `static/css/modern.css` are respected
3. **Error Handling**: If an error occurs in the app, check `logs/` or stdout. The app uses structured logging with logrus
4. **Database Models**: Key models are Integration, IntegrationOutput, Project, User, RequestLog - check `internal/models/` for relationships
5. **API Endpoints**: Most endpoints require authentication except `/webhook/*` and `/metrics/*` - see middleware in routes
6. **Environment Variables**: Always check `.env.example` for available configuration options
7. **Testing**: Use webhook test endpoint `/webhook/test/:token` for integration testing
8. **AI Integration**: When working on AI features, check `internal/ai/` for existing components and `docs/AI_*.md` for specifications
9. **Russian UI**: All user-facing text must be in Russian - check existing templates for consistent terminology
10. **Multiple Outputs**: When working with integrations, remember that each can have multiple outputs via IntegrationOutput model
