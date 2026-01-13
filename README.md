# lucidRAG

Your perfect store assistant - A RAG (Retrieval-Augmented Generation) system built in Go with WhatsApp API integration and Angular administration UI.

## 🏗️ Architecture

lucidRAG follows clean architecture principles with domain-driven design:

- **Backend (Go)**: REST API server with WhatsApp Cloud API integration and RAG capabilities
- **Frontend (Angular)**: Modern admin dashboard for managing knowledge base and monitoring conversations
- **Database**: MongoDB for storing documents, messages, users, and chat sessions

## 📁 Project Structure

```
lucidRAG/
├── cmd/
│   └── api/              # Application entry points
│       └── main.go       # Main server application
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── domain/          # Domain models and interfaces
│   ├── handler/         # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── rag/            # RAG service implementation
│   ├── repository/     # Data persistence layer
│   ├── service/        # Business logic services
│   └── whatsapp/       # WhatsApp API client
├── pkg/                 # Public libraries
│   └── logger/         # Logging utilities
├── admin-ui/           # Angular admin dashboard
│   ├── src/
│   │   ├── app/
│   │   │   ├── components/  # UI components
│   │   │   ├── models/      # TypeScript interfaces
│   │   │   └── services/    # API services
│   │   └── environments/    # Environment configs
│   └── Dockerfile
├── .env.example        # Environment variables template
├── Dockerfile          # Go API Dockerfile
├── docker-compose.yml  # Docker Compose configuration
├── Makefile           # Build automation
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24 or higher
- Node.js 20 or higher
- Docker & Docker Compose (optional)
- MongoDB (if not using Docker)

### Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/elprogramadorgt/lucidRAG.git
cd lucidRAG
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start all services with Docker Compose:
```bash
docker-compose up -d
```

The services will be available at:
- API: http://localhost:8080
- Admin UI: http://localhost:4200
- MongoDB: localhost:27017

### Local Development

#### Backend (Go)

1. Install dependencies:
```bash
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Configure your .env file
```

3. Run the API server:
```bash
make run
# or
go run cmd/api/main.go
```

4. Run tests:
```bash
make test
```

#### Frontend (Angular)

1. Navigate to admin-ui directory:
```bash
cd admin-ui
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm start
```

The Angular app will be available at http://localhost:4200

## 🔧 Configuration

### Environment Variables

Key configuration options in `.env`:

**Server Configuration:**
- `SERVER_HOST`: Server bind address (default: 0.0.0.0)
- `SERVER_PORT`: Server port (default: 8080)
- `ENVIRONMENT`: Environment mode (development/production)

**WhatsApp Configuration:**
- `WHATSAPP_API_KEY`: Your WhatsApp Cloud API access token
- `WHATSAPP_PHONE_NUMBER_ID`: Your WhatsApp phone number ID
- `WHATSAPP_BUSINESS_ACCOUNT_ID`: Your Business Account ID
- `WHATSAPP_WEBHOOK_VERIFY_TOKEN`: Token for webhook verification
- `WHATSAPP_API_VERSION`: API version (default: v17.0)

**RAG Configuration:**
- `RAG_MODEL_NAME`: LLM model name (default: gpt-3.5-turbo)
- `RAG_EMBEDDING_MODEL`: Embedding model (default: text-embedding-ada-002)
- `RAG_CHUNK_SIZE`: Document chunk size (default: 512)
- `RAG_CHUNK_OVERLAP`: Chunk overlap size (default: 50)

**Authentication Configuration:**
- `JWT_SECRET`: Secret key for JWT tokens (min 32 characters)
- `JWT_EXPIRY_HOURS`: Token expiry time in hours (default: 24)
- `OPENAI_API_KEY`: OpenAI API key for embeddings and chat completion

**Database Configuration (MongoDB):**
- `DB_TYPE`: Database type (default: mongodb)
- `DB_HOST`: Database host
- `DB_PORT`: Database port (default: 27017)
- `DB_NAME`: Database name
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password

## 📚 API Documentation

### Health Check
```
GET /healthz              (Liveness check)
GET /readyz               (Readiness check with DB status)
```

### Authentication API
```
POST /api/v1/auth/register   (Register new user)
POST /api/v1/auth/login      (Login and get JWT token)
GET  /api/v1/auth/me         (Get current user - requires auth)
```

### WhatsApp Webhook
```
GET  /api/v1/whatsapp/webhook  (Webhook verification)
POST /api/v1/whatsapp/webhook  (Receive webhook events)
```

### RAG API (requires authentication)
```
POST /api/v1/rag/query   (Query the RAG system)
```

### Documents API (requires admin role)
```
GET    /api/v1/documents           (List documents)
GET    /api/v1/documents?id={id}   (Get document by ID)
POST   /api/v1/documents           (Create document)
PUT    /api/v1/documents           (Update document)
DELETE /api/v1/documents?id={id}   (Delete document)
```

### Conversations API (requires admin role)
```
GET /api/v1/conversations              (List conversations)
GET /api/v1/conversations/{id}         (Get conversation by ID)
GET /api/v1/conversations/{id}/messages (Get conversation messages)
```

## 🎨 Frontend Features

The Angular admin UI provides:

- **Authentication**: Login/Register with JWT-based authentication
- **Dashboard**: Overview of system status and features
- **Document Management**: Upload, edit, and delete knowledge base documents
- **Conversation History**: View WhatsApp conversations and RAG responses
- **Responsive Design**: Works on desktop and mobile devices

## 🧪 Testing

### Go Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

### Angular Tests
```bash
cd admin-ui
npm test
```

## 🛠️ Best Practices & Conventions

### Go Code Style

- **Package Naming**: Use lowercase, single-word names
- **Interfaces**: Named with "-er" suffix or describe behavior
- **Error Handling**: Always check and handle errors explicitly
- **Project Layout**: Follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- **Domain-Driven Design**: Business logic in domain layer, infrastructure concerns separated

### Angular Code Style

- **Components**: Use standalone components (Angular 14+)
- **Services**: Use dependency injection
- **TypeScript**: Strict mode enabled
- **Naming**: 
  - Components: PascalCase (e.g., `Dashboard`)
  - Services: PascalCase with Service suffix (e.g., `DocumentService`)
  - Files: kebab-case (e.g., `document-list.ts`)

### Naming Conventions

**Go:**
- Constants: `PascalCase` or `camelCase`
- Public functions/types: `PascalCase`
- Private functions/types: `camelCase`
- Packages: lowercase

**Angular:**
- Components: `PascalCase`
- Services: `PascalCase`
- Interfaces: `PascalCase`
- Files: `kebab-case`

## 🔐 Security

- Store sensitive credentials in environment variables, never in code
- Use HTTPS in production
- Implement rate limiting for API endpoints
- Validate and sanitize all user inputs
- Keep dependencies updated

## 📦 Building for Production

### Go Binary
```bash
make build
# Binary will be in ./bin/lucidrag
```

### Docker Images
```bash
# Build all images
docker-compose build

# Or build individually
docker build -t lucidrag-api:latest .
cd admin-ui && docker build -t lucidrag-ui:latest .
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 🙋 Support

For questions or issues, please open an issue on GitHub.

## 🗺️ Roadmap

- [x] Implement actual RAG query logic with embeddings
- [x] Add user authentication and authorization
- [x] Implement conversation history view
- [ ] Add support for multiple languages
- [ ] Implement analytics dashboard
- [ ] Add file upload for documents (PDF, DOCX, etc.)
- [ ] Implement vector database integration (Pinecone, Weaviate, etc.)
- [ ] Add webhook event processing logic
- [ ] Implement message queue for async processing
- [ ] Add monitoring and observability (Prometheus, Grafana)
