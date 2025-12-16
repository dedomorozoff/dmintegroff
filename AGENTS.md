# AGENTS.md - Project Context for AI Agents

This file provides high-level context, architectural guidelines, and codebase conventions for AI agents working on **dmIntegroff**.

## Project Overview
**Name**: dmIntegroff
**Description**: A REST API integration management system built with Go. It allows users to configure connections between different APIs, transform data, and automatically transfer it via webhooks.
**Primary Language**: Go (Backend), HTML/JS/CSS (Frontend). 
**UI Language**: Russian (The interface and user-facing documentation are in Russian).

## Technology Stack
- **Backend**: Go 1.20+
- **Web Framework**: Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database**: GORM (`gorm.io/gorm`) supporting SQLite (default) and MySQL.
- **Frontend**: Server-side rendered Go HTML Templates.
- **Styling**: Vanilla CSS ("Modern CSS"), located in `static/css/modern.css`. No heavy CSS frameworks like Tailwind or Bootstrap are currently strictly enforced, mostly custom styles.
- **Scripting**: Vanilla JavaScript, located in `static/js/` and inline in templates.

## Architecture
The project follows a standard Go project layout:

- **`cmd/server/main.go`**: Application entry point.
- **`internal/`**: Core application logic.
    - **`controllers/`**: HTTP request handlers (MVC pattern).
    - **`models/`**: GORM database models.
    - **`routes/`**: Route definitions and grouping.
    - **`services/`**: Business logic encapsulation.
    - **`database/`**: Database connection and initialization.
    - **`logger/`**: Custom logging setup.
- **`templates/`**: HTML templates.
    - **`pages/`**: Full page templates.
    - **`partials/`**: Reusable template fragments (header, sidebar, footer).
- **`static/`**: Publicly accessible assets (CSS, JS, images).
- **`docs/`**: Project documentation (Markdown).
- **`.env`**: Configuration file (Environment variables).

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
- **Integrations**: Source API -> Webhook -> Transformation -> Target API.
- **Artificial Intelligence**: Helper features for creating integrations and mappings.
- **Webhooks**: Handling incoming data, validating JSON, syntax highlighting.
- **Logging**: Request/Response logging, error tracking in `dmintegroff.log`.

## Common Tasks
- **Running the Server**: `go run cmd/server/main.go`
- **Database Migrations**: Handled automatically by GORM on startup (auto-migrate).
- **Configuration**: Managed via `.env` file (e.g., `PORT`, `DB_DSN`). The agent should respect `.env` settings.

## Roadmap & Status
- **Current Focus**: Stability, logging improvements, fixing UI bugs, and enhancing the "AI Context" (this file).
- **Future**: Docker support, GraphQL server mode, visual mapping editor.
- **References**: See `docs/ROADMAP.md` for detailed plans.

## Special Instructions for Agents
1.  **Context Loading**: When starting a task, check `internal/routes/routes.go` to understand the URL structure and `internal/models/` for data structure.
2.  **UI Changes**: When modifying UI, ensure you are editing the correct template in `templates/` and that styles in `static/css/modern.css` are respected.
3.  **Error Handling**: If an error occurs in the app, check `logs/` or stdout. The app uses a custom logger.
