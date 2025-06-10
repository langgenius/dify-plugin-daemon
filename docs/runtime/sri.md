# Dify Plugin Daemon - Serverless Runtime Interface (SRI)

The Serverless Runtime Interface (**SRI**) is a set of HTTP APIs for packaging plugins into serverless components, allowing the Dify Plugin Daemon to remotely launch and operate them on external platforms (e.g., AWS Lambda).

This interface enables the daemon to communicate with remote runtime environments via standard protocols to handle plugin deployment, execution, and instance queries.

> ⚠️ **Note**: This interface is currently in the **Alpha** stage. Stability and backward compatibility are not guaranteed. A production-grade SRI implementation is available in the enterprise edition. For support, please contact `business@dify.ai`.

---

## 🔧 Basic Configuration

The daemon is configured using the following environment variables:

| Variable | Description |
|----------|-------------|
| `DIFY_PLUGIN_SERVERLESS_CONNECTOR_URL` | Base URL of the remote runtime environment, e.g., `https://example.com` |
| `DIFY_PLUGIN_SERVERLESS_CONNECTOR_API_KEY` | Authentication token for accessing SRI, passed in the `Authorization` request header |

---

## 📡 API Endpoints

### `GET /ping`

Used by the daemon for connectivity checks during startup.

**Request**

```http
GET /ping
Authorization: <API_KEY>
```

**Response**

- `200 OK`, response body is plain text: `"pong"`

---

### `GET /v1/runner/instances`

Returns information about plugin instances that are ready to run.

**Query Parameters**

- `filename` (required): Name of the uploaded plugin package, in the format:

  ```
  vendor@plugin@version@hash.difypkg
  ```

**Response**

```json
{
  "items": [
    {
      "ID": "string",
      "Name": "string",
      "Endpoint": "string",
      "ResourceName": "string"
    }
  ]
}
```

---

### `POST /v1/launch`

Launches a plugin using a streaming event protocol for real-time daemon parsing of startup status.

> This API uses `multipart/form-data` for submission and returns status via **Server-Sent Events (SSE)**.

**Request Fields**

| Field      | Type     | Description                                         |
|------------|----------|-----------------------------------------------------|
| `context`  | file     | Plugin package file in `.difypkg` format            |
| `verified` | boolean  | Whether the plugin has been verified by the daemon  |

**SSE Response Format**

```json
{
  "Stage": "healthz|start|build|run|end",
  "State": "running|success|failed",
  "Obj": "string",
  "Message": "string"
}
```

**Stage Descriptions**

| Stage   | Meaning         | Description                                      |
|---------|------------------|--------------------------------------------------|
| healthz | Health check     | Initializes runtime resources and containers     |
| start   | Startup prep     | Prepares the environment                         |
| build   | Build phase      | Builds plugin dependencies and packages image    |
| run     | Execution phase  | Plugin is running; returns key info on success   |
| end     | Completion       | Final state confirmation: success or failure     |

When a message with `Stage=run` and `State=success` is received, the daemon will extract details and register the plugin instance:

```
endpoint=http://...,name=...,id=...
```

**Error Handling**

- If any stage returns `State = failed`, it is considered a launch failure
- The daemon should abort the process and output the `Message` field as the error

---

## 🔁 Communication Sequence (ASCII)

```text
daemon                              Serverless Runtime Interface
   |-------------------------------------->|
   |           GET /ping                  |
   |<--------------------------------------|
   |         200 OK "pong"                |
   |-------------------------------------->|
   |    GET /v1/runner/instances          |
   |            filename                  |
   |<--------------------------------------|
   |             {items}                  |
   |-------------------------------------->|
   |        POST /v1/launch               |
   | context, verified multipart payload |
   |<--------------------------------------|
   |   Building plugin... (SSE)           |
   |<--------------------------------------|
   |   Launching plugin... (SSE)          |
   |<--------------------------------------|
   |   Function: [Name] (SSE)             |
   |<--------------------------------------|
   |   FunctionUrl: [Endpoint] (SSE)      |
   |<--------------------------------------|
   |   Done: Plugin launched (SSE)        |
```

---

## 📦 Plugin File Naming Convention

Plugin files must use the `.difypkg` extension and follow this naming convention:

```
<vendor>@<plugin_name>@<version>@<sha256_hash>.difypkg
```

Example:

```
langgenius@tavily@0.0.5@7f277f7a63e36b1b3e9ed53e55daab0b281599d14902664bade86215f5374f06.difypkg
```

---

## 📬 Contact Us

For access to the enterprise-supported version or more details about plugin packaging and deployment, please contact:

📧 `business@dify.ai`

---

## 📘 OpenAPI Specification (YAML)

```yaml
openapi: 3.0.3
info:
  title: Dify Plugin Daemon - Serverless Runtime Interface (SRI)
  version: alpha
  description: HTTP API specification for the Dify Plugin Daemon's Serverless Runtime
    Interface (SRI).
paths:
  /ping:
    get:
      summary: Health check endpoint
      description: Used by the daemon to verify connectivity with the SRI.
      responses:
        '200':
          description: Returns 'pong' if the service is alive
          content:
            text/plain:
              schema:
                type: string
                example: pong
      security:
      - apiKeyAuth: []
  /v1/runner/instances:
    get:
      summary: List available plugin instances
      parameters:
      - name: filename
        in: query
        required: true
        schema:
          type: string
        description: Full plugin package filename (e.g., vendor@plugin@version@hash.difypkg)
      responses:
        '200':
          description: List of available plugin instances
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: object
                      properties:
                        ID:
                          type: string
                        Name:
                          type: string
                        Endpoint:
                          type: string
                        ResourceName:
                          type: string
      security:
      - apiKeyAuth: []
  /v1/launch:
    post:
      summary: Launch a plugin via SSE
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                context:
                  type: string
                  format: binary
                  description: Plugin package file (.difypkg)
                verified:
                  type: boolean
                  description: Whether the plugin is verified
              required:
              - context
      responses:
        '200':
          description: Server-Sent Events stream with plugin launch stages
          content:
            text/event-stream:
              schema:
                type: object
                properties:
                  Stage:
                    type: string
                    enum:
                    - healthz
                    - start
                    - build
                    - run
                    - end
                  State:
                    type: string
                    enum:
                    - running
                    - success
                    - failed
                  Obj:
                    type: string
                  Message:
                    type: string
      security:
      - apiKeyAuth: []
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
```
