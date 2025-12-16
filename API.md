# API Documentation

This document describes the REST API endpoints available in lucidRAG.

## Base URL

- Development: `http://localhost:8080`
- Production: Configure via `SERVER_HOST` and `SERVER_PORT` environment variables

## Authentication

> **Note**: Authentication is not yet implemented. This section will be updated when authentication is added.

## Endpoints

### Health Check

Check the health status of the API.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "lucidRAG",
  "version": "0.1.0"
}
```

**Status Codes:**
- `200 OK`: Service is healthy

---

### WhatsApp Webhook Verification

Verify webhook subscription with WhatsApp.

**Endpoint:** `GET /webhook/whatsapp`

**Query Parameters:**
- `hub.mode`: Subscription mode (should be "subscribe")
- `hub.verify_token`: Verification token (must match `WHATSAPP_WEBHOOK_VERIFY_TOKEN`)
- `hub.challenge`: Challenge string to return

**Response:**
Returns the challenge string if verification is successful.

**Status Codes:**
- `200 OK`: Verification successful
- `403 Forbidden`: Invalid verification token

---

### WhatsApp Webhook Handler

Receive and process WhatsApp webhook events.

**Endpoint:** `POST /webhook/whatsapp`

**Request Body:**
```json
{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "BUSINESS_ACCOUNT_ID",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "PHONE_NUMBER",
              "phone_number_id": "PHONE_NUMBER_ID"
            },
            "messages": [
              {
                "from": "SENDER_PHONE_NUMBER",
                "id": "MESSAGE_ID",
                "timestamp": "TIMESTAMP",
                "text": {
                  "body": "MESSAGE_CONTENT"
                },
                "type": "text"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

**Response:**
```json
{
  "status": "ok"
}
```

**Status Codes:**
- `200 OK`: Webhook processed successfully
- `400 Bad Request`: Invalid payload
- `500 Internal Server Error`: Processing error

---

### Query RAG System

Send a query to the RAG system to get an intelligent response.

**Endpoint:** `POST /api/v1/rag/query`

**Request Body:**
```json
{
  "query": "What are your store hours?",
  "top_k": 5,
  "threshold": 0.7
}
```

**Parameters:**
- `query` (string, required): The question or query text
- `top_k` (integer, optional): Number of relevant chunks to retrieve (default: 5)
- `threshold` (float, optional): Similarity threshold for chunk retrieval (default: 0.7)

**Response:**
```json
{
  "answer": "Our store is open Monday through Friday from 9 AM to 6 PM...",
  "relevant_chunks": [
    {
      "id": "chunk_123",
      "document_id": "doc_456",
      "chunk_index": 0,
      "content": "Store hours: Monday-Friday 9 AM - 6 PM...",
      "embedding": [],
      "created_at": "2023-12-01T10:00:00Z"
    }
  ],
  "confidence_score": 0.85,
  "processing_time_ms": 234
}
```

**Status Codes:**
- `200 OK`: Query processed successfully
- `400 Bad Request`: Invalid query format
- `500 Internal Server Error`: Processing error

---

### List Documents

Retrieve a list of documents from the knowledge base.

**Endpoint:** `GET /api/v1/documents`

**Query Parameters:**
- `limit` (integer, optional): Maximum number of documents to return (default: 10)
- `offset` (integer, optional): Number of documents to skip (default: 0)

**Response:**
```json
{
  "documents": [
    {
      "id": "doc_123",
      "title": "Store Policies",
      "content": "Full document content...",
      "source": "website",
      "uploaded_at": "2023-12-01T10:00:00Z",
      "updated_at": "2023-12-01T10:00:00Z",
      "is_active": true,
      "metadata": "{}"
    }
  ],
  "limit": 10,
  "offset": 0
}
```

**Status Codes:**
- `200 OK`: Documents retrieved successfully
- `500 Internal Server Error`: Retrieval error

---

### Get Document by ID

Retrieve a specific document by its ID.

**Endpoint:** `GET /api/v1/documents?id={document_id}`

**Query Parameters:**
- `id` (string, required): Document ID

**Response:**
```json
{
  "id": "doc_123",
  "title": "Store Policies",
  "content": "Full document content...",
  "source": "website",
  "uploaded_at": "2023-12-01T10:00:00Z",
  "updated_at": "2023-12-01T10:00:00Z",
  "is_active": true,
  "metadata": "{}"
}
```

**Status Codes:**
- `200 OK`: Document found
- `400 Bad Request`: Missing or invalid ID
- `404 Not Found`: Document not found
- `500 Internal Server Error`: Retrieval error

---

### Create Document

Add a new document to the knowledge base.

**Endpoint:** `POST /api/v1/documents`

**Request Body:**
```json
{
  "title": "Store Policies",
  "content": "Our store policies include...",
  "source": "manual",
  "is_active": true,
  "metadata": "{}"
}
```

**Parameters:**
- `title` (string, required): Document title
- `content` (string, required): Document content
- `source` (string, optional): Source of the document
- `is_active` (boolean, optional): Whether document is active (default: true)
- `metadata` (string, optional): Additional metadata as JSON string

**Response:**
```json
{
  "id": "doc_123",
  "message": "Document added successfully"
}
```

**Status Codes:**
- `201 Created`: Document created successfully
- `400 Bad Request`: Invalid document data
- `500 Internal Server Error`: Creation error

---

### Update Document

Update an existing document in the knowledge base.

**Endpoint:** `PUT /api/v1/documents`

**Request Body:**
```json
{
  "id": "doc_123",
  "title": "Updated Store Policies",
  "content": "Updated content...",
  "source": "manual",
  "is_active": true,
  "metadata": "{}"
}
```

**Parameters:**
- `id` (string, required): Document ID
- `title` (string, optional): Updated document title
- `content` (string, optional): Updated document content
- `source` (string, optional): Updated source
- `is_active` (boolean, optional): Updated active status
- `metadata` (string, optional): Updated metadata

**Response:**
```json
{
  "message": "Document updated successfully"
}
```

**Status Codes:**
- `200 OK`: Document updated successfully
- `400 Bad Request`: Invalid document data
- `404 Not Found`: Document not found
- `500 Internal Server Error`: Update error

---

### Delete Document

Delete a document from the knowledge base.

**Endpoint:** `DELETE /api/v1/documents?id={document_id}`

**Query Parameters:**
- `id` (string, required): Document ID

**Response:**
```json
{
  "message": "Document deleted successfully"
}
```

**Status Codes:**
- `200 OK`: Document deleted successfully
- `400 Bad Request`: Missing or invalid ID
- `404 Not Found`: Document not found
- `500 Internal Server Error`: Deletion error

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common HTTP status codes:
- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: HTTP method not supported
- `500 Internal Server Error`: Server error

## Rate Limiting

> **Note**: Rate limiting is not yet implemented. This section will be updated when rate limiting is added.

## CORS

The API supports Cross-Origin Resource Sharing (CORS) with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

## Examples

### Using cURL

**Query RAG System:**
```bash
curl -X POST http://localhost:8080/api/v1/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What are your store hours?",
    "top_k": 5,
    "threshold": 0.7
  }'
```

**Create Document:**
```bash
curl -X POST http://localhost:8080/api/v1/documents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Store Hours",
    "content": "We are open Monday-Friday 9AM-6PM",
    "source": "manual",
    "is_active": true
  }'
```

**List Documents:**
```bash
curl http://localhost:8080/api/v1/documents?limit=10&offset=0
```

### Using JavaScript/TypeScript

```typescript
// Query RAG
const response = await fetch('http://localhost:8080/api/v1/rag/query', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    query: 'What are your store hours?',
    top_k: 5,
    threshold: 0.7
  })
});

const data = await response.json();
console.log(data.answer);

// Create Document
const createResponse = await fetch('http://localhost:8080/api/v1/documents', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: 'Store Hours',
    content: 'We are open Monday-Friday 9AM-6PM',
    source: 'manual',
    is_active: true
  })
});

const result = await createResponse.json();
console.log(result.id);
```

---

### List Conversations

Retrieve all WhatsApp conversation sessions.

**Endpoint:** `GET /api/v1/conversations`

**Query Parameters:**
- `limit` (integer, optional): Maximum number of sessions to return (default: 50)
- `offset` (integer, optional): Number of sessions to skip (default: 0)

**Response:**
```json
{
  "sessions": [
    {
      "id": "session_1",
      "user_phone_number": "+1234567890",
      "started_at": "2023-12-01T10:00:00Z",
      "last_message_at": "2023-12-01T15:30:00Z",
      "is_active": true,
      "context": ""
    }
  ],
  "limit": 50,
  "offset": 0
}
```

**Status Codes:**
- `200 OK`: Sessions retrieved successfully
- `500 Internal Server Error`: Retrieval error

---

### Get Conversation Session

Retrieve a specific conversation session by ID.

**Endpoint:** `GET /api/v1/conversations/session?id={session_id}`

**Query Parameters:**
- `id` (string, required): Session ID

**Response:**
```json
{
  "id": "session_1",
  "user_phone_number": "+1234567890",
  "started_at": "2023-12-01T10:00:00Z",
  "last_message_at": "2023-12-01T15:30:00Z",
  "is_active": true,
  "context": ""
}
```

**Status Codes:**
- `200 OK`: Session found
- `400 Bad Request`: Missing session ID
- `404 Not Found`: Session not found
- `500 Internal Server Error`: Retrieval error

---

### Get Conversation Messages

Retrieve messages for a specific conversation session.

**Endpoint:** `GET /api/v1/conversations/messages?session_id={session_id}`

**Query Parameters:**
- `session_id` (string, required): Session ID
- `limit` (integer, optional): Maximum number of messages to return (default: 100)
- `offset` (integer, optional): Number of messages to skip (default: 0)

**Response:**
```json
{
  "messages": [
    {
      "id": "msg_123",
      "from": "+1234567890",
      "to": "PHONE_NUMBER_ID",
      "content": "Hello, I have a question",
      "message_type": "text",
      "timestamp": "2023-12-01T10:00:00Z",
      "status": "received"
    }
  ],
  "session_id": "session_1",
  "limit": 100,
  "offset": 0
}
```

**Status Codes:**
- `200 OK`: Messages retrieved successfully
- `400 Bad Request`: Missing session ID
- `500 Internal Server Error`: Retrieval error

---

## WebSocket Support

> **Note**: WebSocket support is not yet implemented. This section will be updated when WebSocket functionality is added for real-time updates.

## Versioning

The API uses URL versioning. Current version is `v1` as indicated in the URL path: `/api/v1/...`

When breaking changes are introduced, a new version will be released (e.g., `/api/v2/...`).
