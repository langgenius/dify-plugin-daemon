# Database Operations

**Related Documentation:**
- [Generic Types](generics.md) - Type-safe database queries
- [Cache Operations](cache.md) - Database result caching patterns

## Configuration

Configure PostgreSQL/MySQL in `.env`:
```bash
DB_TYPE=postgresql  # or mysql
DB_HOST=localhost
DB_DATABASE=dify_plugin
DB_USERNAME=postgres
DB_PASSWORD=password
```

## Query Builder Pattern

GORM with custom query builder (`internal/db/`):

### Basic CRUD
```go
db.Create(&model)
db.Update(&model)
db.Delete(&model)
db.DeleteByCondition(condition)
```

### Type-safe Queries
```go
// Get single record
plugin, _ := db.GetOne[models.Plugin](
    db.Equal("id", pluginID),
    db.Equal("tenant_id", tenantID),
)

// Get multiple records
plugins, _ := db.GetAll[models.Plugin](
    db.Equal("tenant_id", tenantID),
    db.Page(1, 20),
    db.OrderBy("created_at", true),
)

// Count records
count, _ := db.GetCount[models.Plugin](
    db.Equal("tenant_id", tenantID),
)
```

### Transactions
```go
db.WithTransaction(func(tx *gorm.DB) error {
    if err := db.Create(&plugin, tx); err != nil {
        return err
    }
    return db.Update(&installation, tx)
})
```

## Query Functions

- `Equal`, `NotEqual`: Basic comparisons
- `GreaterThan`, `LessThan`: Numeric comparisons  
- `Like`: Pattern matching
- `Page(page, pageSize)`: Pagination
- `OrderBy(field, desc)`: Sorting
- `Preload(relation)`: Eager loading
- `WithTransaction`: Transaction wrapper
- `WLock`: Write lock for updates

## Models

All database models in `internal/types/models/`:
- `Plugin`, `PluginInstallation`, `PluginDeclaration`
- `ToolInstallation`, `AIModelInstallation`, `AgentStrategyInstallation`
- `Endpoint`, `ServerlessRuntime`, `TenantStorage`

## Plugin Declaration

```go
// Get plugin with remote declaration
plugin, _ := db.GetOne[models.Plugin](
    db.Equal("plugin_id", pluginID),
)
declaration := plugin.RemoteDeclaration

// Get stored declaration
decl, _ := db.GetOne[models.PluginDeclaration](
    db.Equal("plugin_id", pluginID),
)
```