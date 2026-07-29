# Dify Plugin Daemon - Serverless Runtime Interface (SRI)

Serverless Runtime Interface (**SRI**) 是一组用于将插件封装为 Serverless 组件，并由 Dify Plugin Daemon 在外部平台（如 AWS Lambda）上远程启动和运行的 HTTP 接口规范。

该接口允许 daemon 通过标准协议与远程运行环境通信，实现插件部署、运行、实例查询等功能。

> ⚠️ **注意**：当前接口处于 **Alpha 阶段**，不保证稳定性与向后兼容性。 企业版中提供生产级 SRI 实现, 如需请联系 `business@dify.ai`。

---

## 🔧 基础配置

daemon 通过如下环境变量进行配置：

| 变量名 | 含义 |
|--------|------|
| `DIFY_PLUGIN_SERVERLESS_CONNECTOR_URL` | 指定远程运行环境的 Base URL，例如 `https://example.com` |
| `DIFY_PLUGIN_SERVERLESS_CONNECTOR_API_KEY` | 用于访问 SRI 的鉴权 token，将被加入请求 Header 中的 `Authorization` 字段 |
| `MAX_SERVERLESS_REQUEST_BYTES` | 发送到 serverless 插件运行时的最大序列化请求负载大小（以字节为单位）。默认值为 5242880（5 MB）。此限制考虑了 Lambda Function URL 的 6 MB 请求大小限制，并为 headers 和元数据留有安全余量。 |

---

## 📏 请求大小限制与批处理

像 AWS Lambda Function URL 这样的 serverless 平台有严格的负载大小限制（Lambda Function URL 为 6 MB）。

当文本嵌入和 token 计数请求超过 `MAX_SERVERLESS_REQUEST_BYTES` 时，daemon 会自动将其拆分为多个批次：

- **透明处理**：批处理对插件开发者是透明的 - daemon 自动处理拆分和合并
- **顺序处理**：每个批次按顺序处理以确保结果一致
- **自动合并**：来自所有批次的结果会自动合并为单个响应
- **错误处理**：如果单个文本项超过限制，请求将失败并返回 `ServerlessPayloadTooLargeError` 错误

此功能可防止大型嵌入请求因负载大小限制而失败时出现不明确的错误。

---

## 📡 接口说明

### `GET /ping`

用于 daemon 启动时的连通性检查。

**请求**

```http
GET /ping
Authorization: <API_KEY>
```

**响应**

- `200 OK`，响应体为纯文本字符串 `"pong"`

---

### `GET /v1/runner/instances`

返回支持运行的插件实例信息。

**请求参数**

- `filename`（必填）：上传的插件包文件名，格式为：

  ```
  vendor@plugin@version@hash.difypkg
  ```

**响应**

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

以流式事件的方式启动插件，供 daemon 实时解析启动状态。

> 本接口使用 `multipart/form-data` 提交，同时以 **Server-Sent Events（SSE）** 返回插件运行状态流。

**请求字段**

| 字段名     | 类型      | 描述                                         |
|------------|-----------|----------------------------------------------|
| `context`  | file      | `.difypkg` 格式的插件包                      |
| `verified` | boolean   | 插件是否已通过 daemon 验证（由 manifest 判断） |

**SSE 响应格式**

```json
{
  "Stage": "healthz|start|build|run|end",
  "State": "running|success|failed",
  "Obj": "string",
  "Message": "string"
}
```

**阶段说明**

| Stage   | 含义         | 行为说明                                       |
|---------|--------------|------------------------------------------------|
| healthz | 健康检查     | 初始化运行时资源，准备插件容器                |
| start   | 启动准备阶段 | 准备环境                                      |
| build   | 构建阶段     | 构建插件依赖，打包镜像                        |
| run     | 运行阶段     | 插件运行中，如成功将返回关键信息              |
| end     | 启动完成     | 插件运行结果确认，可能为 success 或 failed     |

当接收到以下格式的 `Stage=run` 且 `State=success` 消息时，daemon 将提取其中信息并建立插件实例：

```
endpoint=http://...,name=...,id=...
```

**错误处理**

- 任意阶段返回 `State = failed` 即视为启动失败
- daemon 应中断流程并抛出异常，输出 `Message` 内容作为错误信息

---

## 🔁 通信时序图（ASCII）

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

## 📦 插件文件名规范

插件文件扩展名必须为 `.difypkg`，命名格式如下：

```
<vendor>@<plugin_name>@<version>@<sha256_hash>.difypkg
```

示例：

```
langgenius@tavily@0.0.5@7f277f7a63e36b1b3e9ed53e55daab0b281599d14902664bade86215f5374f06.difypkg
```

---

## 📬 联系我们

如需接入商业支持版本，或希望深入了解插件打包与部署规范，请联系：

📧 `business@dify.ai`

---

## 📘 OpenAPI 规范（YAML）

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
