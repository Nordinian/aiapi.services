# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**New API** is a comprehensive AI API gateway and asset management system built on Go + React. It acts as a unified proxy layer for 30+ AI providers (OpenAI, Claude, Gemini, Baidu, etc.), providing authentication, rate limiting, cost tracking, and administrative controls. This is a fork of One API with significant enhancements.

## Development Commands

### Backend (Go)
```bash
# Start development server
go run main.go

# Build production binary  
go build -o new-api main.go

# Run with custom port and log directory
go run main.go --port 3001 --log-dir ./custom-logs

# Database migrations are automatic on startup
```

### Frontend (React + Vite)
```bash
cd web

# Development server with hot reload
npm run dev
# or
bun dev

# Production build
npm run build
# or  
bun run build

# Code formatting
npm run lint:fix
```

### Full Stack Development
```bash
# Build frontend and start backend (from project root)
make all

# Frontend only
make build-frontend

# Backend only  
make start-backend
```

### Docker Development
```bash
# Full stack with MySQL + Redis
docker-compose up -d

# View logs
docker-compose logs -f new-api

# Rebuild after changes
docker-compose up --build -d
```

## High-Level Architecture

### System Design Pattern
**Layered Architecture + API Gateway Pattern** with extensive adapter system for AI provider integration.

### Core Components

**Backend (Go) - `/aiapi.services/`**
- **`main.go`** - Application entry point with embedded frontend
- **`router/`** - HTTP routing (API, relay, web, dashboard routes)
- **`controller/`** - Request handlers and business logic
- **`middleware/`** - Cross-cutting concerns (auth, rate limiting, logging, CORS)
- **`model/`** - GORM data models with auto-migration support
- **`service/`** - Business services and utility functions
- **`relay/`** - **Core AI integration layer** with provider-specific adapters
- **`setting/`** - Configuration management and system settings
- **`common/`** - Shared utilities, constants, and initialization

**Frontend (React) - `/aiapi.services/web/src/`**
- **`pages/`** - Route-level components (Dashboard, Settings, Playground, etc.)
- **`components/`** - Reusable UI components organized by feature
- **`context/`** - React Context providers for global state (User, Status, Theme)
- **`hooks/`** - Custom hooks for API integration and state management
- **`helpers/`** - API client and utility functions

### AI Relay System Architecture

The **`relay/`** directory is the heart of the system - a sophisticated adapter pattern implementation:

```
relay/
├── channel/                 # Provider-specific adapters (30+ AI services)
│   ├── openai/             # OpenAI GPT models
│   ├── claude/             # Anthropic Claude  
│   ├── gemini/             # Google Gemini
│   ├── aws/                # AWS Bedrock
│   ├── azure/              # Azure OpenAI
│   ├── baidu/              # Baidu ERNIE
│   ├── zhipu/              # Zhipu AI (ChatGLM)
│   └── [25+ more providers]
├── common/                 # Shared relay utilities
└── *_handler.go            # Request type handlers (text, image, audio, etc.)
```

**Adapter Interface Pattern:**
Each provider implements a standard interface:
- `ConvertOpenAIRequest()` - Transform OpenAI format to provider format
- `DoRequest()` - Execute provider HTTP request  
- `DoResponse()` - Transform provider response back to OpenAI format
- Provider-specific authentication and endpoint handling

### Data Flow
1. **Request** → Authentication middleware → Rate limiting → Channel selection
2. **Transform** → OpenAI format → Provider-specific format via adapter
3. **Execute** → HTTP request to AI provider
4. **Response** → Provider response → OpenAI format → Client
5. **Logging** → Usage tracking, billing, analytics

### Database Schema
- **Users** - Authentication and authorization with role hierarchy
- **Tokens** - API key management with usage quotas
- **Channels** - AI provider configurations and health status
- **Logs** - Detailed request/response tracking for billing
- **Tasks** - Async operations (Midjourney, Suno, etc.)
- **Pricing** - Cost management and billing calculations

### Configuration System
- **Environment variables** (see `.env.example`) for deployment settings
- **Database-stored options** via `setting/` package for runtime configuration
- **Dynamic channel management** with health checking and auto-disable
- **Multi-node support** with master/worker configuration

## Key Features Understanding

### Authentication & Authorization
- **Multi-level roles**: Guest, User, Admin, Root
- **Session-based auth** for web dashboard  
- **API key auth** for programmatic access
- **OAuth integration** (GitHub, Google, LinuxDO, Telegram, OIDC)

### Rate Limiting & Quotas
- **Global rate limits** to protect system resources
- **User quotas** with credit-based billing
- **Model-specific limits** for different AI services
- **Channel-level controls** for provider management

### Cost Management
- **Real-time usage tracking** with detailed logging
- **Flexible pricing models** (per-token, per-request, subscription)
- **Credit system** with top-up functionality
- **Detailed analytics** for usage patterns

### Advanced Features
- **Streaming responses** via Server-Sent Events
- **Request caching** with Redis backend
- **Health monitoring** with automatic channel failover
- **Multi-language support** (English/Chinese)
- **Playground interface** for model testing and debugging

## Configuration Files

### Environment Configuration
- **`.env.example`** - Comprehensive environment variable reference
- **`docker-compose.yml`** - Full-stack deployment with MySQL + Redis
- **`cloudbuild.yaml`** - Google Cloud Build configuration
- **`one-api.service`** - Systemd service configuration

### Frontend Build Configuration  
- **`web/vite.config.js`** - Vite build optimization with manual chunking
- **`web/tailwind.config.js`** - TailwindCSS configuration
- **`web/package.json`** - React dependencies and Semi Design components

## Common Development Patterns

### Adding New AI Provider
1. Create adapter in `relay/channel/[provider]/` with required interface
2. Add provider constants in `constant/` 
3. Register adapter in channel factory
4. Update frontend provider options
5. Add provider-specific settings if needed

### Database Model Changes
- GORM auto-migration handles schema updates automatically
- Add migration logic in `model/main.go` if needed for data transformations
- Test migrations with different database backends (SQLite, MySQL, PostgreSQL)

### API Endpoint Development
1. Define routes in appropriate `router/*.go` file
2. Implement handler in `controller/` with request validation
3. Add middleware if needed (auth, rate limiting, etc.)
4. Update frontend API client in `web/src/helpers/api.js`

## Testing Strategy

### Backend Testing
- Use Go's built-in testing framework
- Test adapter interfaces with mock providers
- Database testing with test databases
- Integration testing with real AI providers (use test models)

### Frontend Testing
- Component testing with React Testing Library
- API integration testing with mock responses  
- E2E testing for critical user workflows
- Playground functionality testing with various models

## Security Considerations

- **Never commit secrets** - Use environment variables or secure vaults
- **API key rotation** - Implement regular rotation for provider keys
- **Rate limiting enforcement** - Prevent abuse and resource exhaustion
- **Input validation** - Sanitize all user inputs and API requests
- **Audit logging** - Track all administrative actions
- **CORS configuration** - Properly configure for production domains