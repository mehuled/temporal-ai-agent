# Temporal AI Agent

A Go application that demonstrates Temporal workflows and activities with an HTTP API server.

## Setup

1. Copy the environment template:
   ```bash
   cp .env.example .env
   ```

2. Update the `.env` file with your Temporal configuration:
   - `TEMPORAL_HOST_PORT`: Your Temporal server host and port
   - `TEMPORAL_NAMESPACE`: Your Temporal namespace
   - `TEMPORAL_API_KEY`: Your Temporal API key
   - `TEMPORAL_TASK_QUEUE`: The task queue name
   - `TEMPORAL_TLS_ENABLED`: Set to `true` for production
   - `SERVER_PORT`: API server port (default: 3000)

## Running the Application

### Start the Worker
```bash
go run ./worker
```

### Start the API Server
```bash
go run ./api
```

The API server will start on port 3000 (or the port specified in `SERVER_PORT`).

## API Endpoints

### POST /start-workflow
Starts a new workflow with the provided message.

**Request:**
```json
{
  "message": "Hello World"
}
```

**Response:**
```json
{
  "workflow_id": "chat-workflow-1234567890",
  "run_id": "0199bb76-5886-7ba4-bcc8-2e02eb138729",
  "result": "Hello Hello World!"
}
```

### POST /signal/user-prompt
Sends a user prompt signal to an existing workflow.

**Request:**
```json
{
  "workflow_id": "chat-workflow-1234567890",
  "run_id": "0199bb76-5886-7ba4-bcc8-2e02eb138729",
  "message": "What's the weather like?"
}
```

**Response:**
```json
{
  "success": true
}
```

### POST /signal/confirm
Sends a confirmation signal to an existing workflow.

**Request:**
```json
{
  "workflow_id": "chat-workflow-1234567890",
  "run_id": "0199bb76-5886-7ba4-bcc8-2e02eb138729",
  "message": "Yes, I confirm"
}
```

**Response:**
```json
{
  "success": true
}
```

### POST /signal/end-chat
Sends an end chat signal to terminate a workflow.

**Request:**
```json
{
  "workflow_id": "chat-workflow-1234567890",
  "run_id": "0199bb76-5886-7ba4-bcc8-2e02eb138729",
  "message": "Goodbye"
}
```

**Response:**
```json
{
  "success": true
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

## Environment Variables

All configuration is loaded from environment variables, with the following defaults:

- `TEMPORAL_HOST_PORT`: `localhost:7233`
- `TEMPORAL_NAMESPACE`: `default`
- `TEMPORAL_API_KEY`: (required, no default)
- `TEMPORAL_TASK_QUEUE`: `my-task-queue`
- `TEMPORAL_TLS_ENABLED`: `false`
- `SERVER_PORT`: `3000`

The application will first try to load variables from a `.env` file, then fall back to system environment variables.

## Example Usage

1. Start the worker:
   ```bash
   go run ./worker
   ```

2. Start the API server:
   ```bash
   go run ./api
   ```

3. Send a workflow request:
   ```bash
   curl -X POST http://localhost:3000/start-workflow \
     -H "Content-Type: application/json" \
     -d '{"message": "Hello from API!"}'
   ```

4. Send signals to the workflow:
   ```bash
   # Send user prompt signal
   curl -X POST http://localhost:3000/signal/user-prompt \
     -H "Content-Type: application/json" \
     -d '{"workflow_id": "YOUR_WORKFLOW_ID", "message": "What can you help me with?"}'
   
   # Send confirmation signal
   curl -X POST http://localhost:3000/signal/confirm \
     -H "Content-Type: application/json" \
     -d '{"workflow_id": "YOUR_WORKFLOW_ID", "message": "Yes, I understand"}'
   
   # End the chat
   curl -X POST http://localhost:3000/signal/end-chat \
     -H "Content-Type: application/json" \
     -d '{"workflow_id": "YOUR_WORKFLOW_ID", "message": "Thank you, goodbye!"}'
   ```