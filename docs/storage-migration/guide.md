# Storage Migration Guide (Local → Cloud)

This guide explains how to use the migration CLI to copy your local plugin storage to a cloud object storage (S3, COS, OSS, Azure Blob, GCS, OBS, TOS).

- Prerequisites
  - A target bucket/container that already exists and grants write access.
  - Cloud storage configuration is provided via environment variables or `.env` (same variables as the daemon).
  - Only “local → cloud” is supported; “local → local” is not allowed.

- Required environment variables (examples)
  - Basics
    - `PLUGIN_STORAGE_TYPE`: Target storage type, e.g., `s3`, `tencent` (COS), `aliyun_oss`, `azure_blob`, `gcs`, `huawei_obs`, `volcengine_tos`
    - `PLUGIN_STORAGE_OSS_BUCKET`: Target bucket/container name
    - `PLUGIN_STORAGE_LOCAL_ROOT`: Local storage root, default `./storage`
  - Provider credentials (as needed)
    - AWS S3: `AWS_ACCESS_KEY`, `AWS_SECRET_KEY`, `AWS_REGION`, `S3_ENDPOINT` (optional), `S3_USE_PATH_STYLE`, `S3_USE_AWS`
    - Tencent COS: `TENCENT_COS_SECRET_ID`, `TENCENT_COS_SECRET_KEY`, `TENCENT_COS_REGION`
    - Other providers: see fields in `internal/server/server.go`.

- What gets migrated
  - `plugin_packages`: Plugin package cache
  - `assets`: Plugin media/icons cache
  - `plugin`: Installed plugin archives

- How to run
  - Direct (reads `.env`)
    - `go run ./cmd/migrate_storage --dry-run` to preview
    - `go run ./cmd/migrate_storage` to execute
  - Build a binary
    - `go build -o migrate-storage ./cmd/migrate_storage`
    - `./migrate-storage --only packages,assets,installed`

- Useful flags
  - `--dry-run`: Print planned copies without uploading
  - `--only`: Limit scope (comma-separated): `packages,assets,installed`
  - `--source-root`: Override local storage root (default from `PLUGIN_STORAGE_LOCAL_ROOT`)

- Behavior
  - Idempotent: existing destination objects are skipped; safe to rerun
  - Restriction: if `PLUGIN_STORAGE_TYPE=local`, the tool exits (local → cloud only)

- Troubleshooting
  - DNS/network errors: check connectivity, proxy, or private network policies
  - Access denied: verify AccessKey/Secret, IAM/STS, container permissions, and bucket existence
  - Local read failures: ensure `PLUGIN_STORAGE_LOCAL_ROOT` points to the correct directory structure

- Directory layout reference
  - Expected subdirectories under local root:
    - `plugin_packages/`
    - `assets/`
    - `plugin/`


