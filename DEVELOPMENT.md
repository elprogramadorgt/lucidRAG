# Development Guide

This guide provides detailed information for developers working on lucidRAG.

## Architecture Overview

lucidRAG follows clean architecture principles:

```
┌─────────────────┐
│   HTTP Layer    │ (handlers, middleware)
└────────┬────────┘
         │
┌────────▼────────┐
│  Service Layer  │ (business logic)
└────────┬────────┘
         │
┌────────▼────────┐
│  Domain Layer   │ (models, interfaces)
└────────┬────────┘
         │
┌────────▼────────┐
│   Repository    │ (data access)
└─────────────────┘
```

### Key Principles

1. **Dependency Inversion**: Inner layers don't depend on outer layers
2. **Separation of Concerns**: Each layer has a specific responsibility
3. **Testability**: Business logic is independent of frameworks
4. **Interface-Based Design**: Use interfaces for loose coupling

## Project Structure Explained

### Backend (Go)

- **cmd/**: Application entry points (main.go files)
- **internal/**: Private application code
  - **config/**: Configuration management
  - **domain/**: Core business models and interfaces
  - **handler/**: HTTP request handlers
  - **middleware/**: HTTP middleware
  - **rag/**: RAG implementation
  - **repository/**: Data persistence
  - **service/**: Business logic
  - **whatsapp/**: WhatsApp API client
- **pkg/**: Public reusable packages

### Frontend (Angular)

- **src/app/components/**: UI components
- **src/app/services/**: API client services
- **src/app/models/**: TypeScript interfaces
- **src/environments/**: Environment configurations

## Development Workflow

### 1. Starting the Development Environment

**Backend:**
```bash
# Terminal 1: Start PostgreSQL
docker-compose up postgres

# Terminal 2: Run API server
make run
```

**Frontend:**
```bash
# Terminal 3: Start Angular dev server
cd admin-ui
npm start
```

### 2. Making Changes

1. Create a feature branch
2. Make your changes
3. Run tests locally
4. Commit with descriptive message
5. Push and create PR

### 3. Testing Changes

**Unit Tests:**
```bash
# Go
make test

# Angular
cd admin-ui && npm test
```

**Manual Testing:**
- Test API endpoints with curl or Postman
- Test UI in browser at http://localhost:4200
- Verify Docker build works

## Common Development Tasks

### Adding a New API Endpoint

1. Define domain model in `internal/domain/models.go`
2. Add interface method in `internal/domain/interfaces.go`
3. Implement service in appropriate service package
4. Create handler in `internal/handler/`
5. Register route in `cmd/api/main.go`
6. Add tests

Example:
```go
// 1. Domain model
type User struct {
    ID    string
    Name  string
    Email string
}

// 2. Interface
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

// 3. Service implementation
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    // Implementation
}

// 4. Handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Handler logic
}

// 5. Route registration
mux.HandleFunc("/api/v1/users", userHandler.GetUser)
```

### Adding a New Angular Component

1. Generate component: `ng generate component components/my-component`
2. Implement component logic
3. Add routing if needed
4. Create service for API calls
5. Add styling
6. Write tests

### Implementing RAG Logic

The RAG service (`internal/rag/service.go`) has placeholder methods that need implementation:

1. **Document Chunking**: Split documents into chunks
2. **Embedding Generation**: Generate embeddings for chunks
3. **Vector Search**: Find similar chunks based on query
4. **Response Generation**: Use LLM to generate answers

Example integration points:
- OpenAI API for embeddings and generation
- Vector databases (Pinecone, Weaviate, etc.)
- Local embedding models

### WhatsApp Integration

The WhatsApp client (`internal/whatsapp/client.go`) needs:

1. **Webhook Processing**: Parse and handle incoming messages
2. **Message Routing**: Route messages to RAG service
3. **Response Formatting**: Format RAG responses for WhatsApp
4. **Error Handling**: Handle API errors gracefully

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Key variables to configure:
- WhatsApp API credentials
- Database connection
- RAG model settings

## Debugging

### Go Debugging

Use delve debugger:
```bash
dlv debug ./cmd/api
```

Or use VS Code launch configuration:
```json
{
  "name": "Launch API",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}/cmd/api"
}
```

### Angular Debugging

Use browser DevTools:
- Set breakpoints in source tab
- Use Angular DevTools extension
- Check network tab for API calls

## Performance Considerations

### Backend
- Use connection pooling for database
- Implement caching for frequently accessed data
- Use goroutines for concurrent operations
- Profile with pprof when needed

### Frontend
- Use OnPush change detection strategy
- Lazy load routes
- Optimize bundle size
- Use trackBy in *ngFor

## Security Best Practices

1. **Never commit secrets** - Use environment variables
2. **Validate all inputs** - Sanitize user data
3. **Use HTTPS** in production
4. **Implement rate limiting** - Prevent abuse
5. **Keep dependencies updated** - Security patches
6. **Use prepared statements** - Prevent SQL injection
7. **Implement proper CORS** - Restrict origins

## Useful Commands

### Go
```bash
make build          # Build binary
make run           # Run application
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make clean         # Clean build artifacts
```

### Angular
```bash
npm start          # Start dev server
npm run build      # Build for production
npm test           # Run tests
npm run lint       # Run linter
```

### Docker
```bash
docker-compose up                    # Start all services
docker-compose up -d                 # Start in background
docker-compose down                  # Stop services
docker-compose logs -f api           # View API logs
docker-compose exec api /bin/sh      # Shell into API container
```

## Troubleshooting

### Go Build Issues

**Problem**: Module import errors
```bash
go mod tidy
go mod download
```

**Problem**: Compilation errors
- Check Go version (requires 1.24+)
- Clear module cache: `go clean -modcache`

### Angular Build Issues

**Problem**: Module not found
```bash
rm -rf node_modules package-lock.json
npm install
```

**Problem**: Type errors
- Check TypeScript version
- Run `npm run build` for detailed errors

### Docker Issues

**Problem**: Container won't start
```bash
docker-compose logs <service-name>
docker-compose down -v  # Remove volumes
docker-compose up --build
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Angular Documentation](https://angular.io/docs)
- [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api)
- [RAG Concepts](https://www.promptingguide.ai/techniques/rag)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
