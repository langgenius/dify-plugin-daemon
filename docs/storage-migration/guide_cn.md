# 存储迁移使用教程（本地 → 云）

本教程介绍如何使用迁移 CLI 将本地插件存储迁移到云对象存储（S3、COS、OSS、Azure Blob、GCS、OBS、TOS）。

- 前提条件
  - 已存在可用的目标存储桶/容器，并具备写权限。
  - 在环境变量或 `.env` 中正确配置云存储信息（与守护进程相同的变量）。
  - 当前仅支持“本地 → 云”，不支持“本地 → 本地”。

- 需要的环境变量（示例）
  - 基本
    - `PLUGIN_STORAGE_TYPE`: 目标存储类型，如 `s3`、`tencent`（COS）、`aliyun_oss`、`azure_blob`、`gcs`、`huawei_obs`、`volcengine_tos`
    - `PLUGIN_STORAGE_OSS_BUCKET`: 目标桶/容器名
    - `PLUGIN_STORAGE_LOCAL_ROOT`: 本地存储根目录，默认 `./storage`
  - 云厂商凭证（按需）
    - AWS S3: `AWS_ACCESS_KEY`、`AWS_SECRET_KEY`、`AWS_REGION`、`S3_ENDPOINT`（可选）、`S3_USE_PATH_STYLE`、`S3_USE_AWS`
    - 腾讯云 COS: `TENCENT_COS_SECRET_ID`、`TENCENT_COS_SECRET_KEY`、`TENCENT_COS_REGION`
    - 其他云参见 `internal/server/server.go` 对应字段。

- 迁移内容
  - `plugin_packages`：插件包缓存
  - `assets`：插件媒体/图标缓存
  - `plugin`：已安装插件归档

- 运行方式
  - 直接运行（读取 `.env`）
    - `go run ./cmd/migrate_storage --dry-run` 先预览
    - `go run ./cmd/migrate_storage` 正式迁移
  - 构建可执行文件
    - `go build -o migrate-storage ./cmd/migrate_storage`
    - `./migrate-storage --only packages,assets,installed`

- 常用参数
  - `--dry-run`：仅打印将要复制的对象，不实际上传
  - `--only`：限制迁移范围，逗号分隔：`packages,assets,installed`
  - `--source-root`：覆盖本地存储根（默认取 `PLUGIN_STORAGE_LOCAL_ROOT`）

- 行为说明
  - 幂等：目标端已存在的对象会跳过，可多次执行
  - 限制：若 `PLUGIN_STORAGE_TYPE=local`，程序将直接退出（仅支持本地 → 云）

- 排障指引
  - DNS 或网络错误：检查本机网络、代理或云厂商私网策略
  - 权限拒绝：确认 AccessKey/Secret、IAM/STS、容器权限、桶/容器是否存在
  - 读取失败（本地文件不存在）：确认 `PLUGIN_STORAGE_LOCAL_ROOT` 指向正确存储目录结构

- 文件结构参考
  - 本地根目录下的关键子目录：
    - `plugin_packages/`
    - `assets/`
    - `plugin/`

