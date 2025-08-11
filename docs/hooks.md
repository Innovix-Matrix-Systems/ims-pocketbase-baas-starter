# Hooks Module

This module provides a structured way to register and manage PocketBase event hooks in your application.

## Overview

The hooks module follows the same pattern as the routes module, organizing event hooks into logical groups and providing dedicated handlers for different types of events.

## Structure

```
internal/
├── hooks/
│   ├── hooks.go          # Main hooks registration
└── handlers/
    └── hook/
        ├── record_hooks.go     # Record model event handlers
        ├── collection_hooks.go # Collection model event handlers
        ├── request_hooks.go    # Request event handlers
        ├── mailer_hooks.go     # Mailer event handlers
        └── realtime_hooks.go   # Realtime event handlers
```

## Usage

### 1. Register Hooks in Your App

In your main application file (typically `internal/app/app.go`), call the hooks registration:

```go
import "ims-pocketbase-baas-starter/internal/hooks"

func NewApp() *pocketbase.PocketBase {
    app := pocketbase.New()
    
    // Register hooks
    hooks.RegisterHooks(app)
    
    return app
}
```

### 2. Available Hook Categories

#### Record Hooks
- `OnRecordCreate` - Triggered when records are created
- `OnRecordUpdate` - Triggered when records are updated
- `OnRecordDelete` - Triggered when records are deleted
- Collection-specific variants available

#### Collection Hooks
- `OnCollectionCreate` - Triggered when collections are created
- `OnCollectionUpdate` - Triggered when collections are updated
- `OnCollectionDelete` - Triggered when collections are deleted

#### Request Hooks
- `OnRecordListRequest` - Triggered on record list API requests
- `OnRecordViewRequest` - Triggered on record view API requests
- `OnRecordCreateRequest` - Triggered on record create API requests
- `OnRecordUpdateRequest` - Triggered on record update API requests
- `OnRecordDeleteRequest` - Triggered on record delete API requests

#### Mailer Hooks
- `OnMailerSend` - Triggered when emails are sent

#### Realtime Hooks
- `OnRealtimeConnect` - Triggered when realtime clients connect
- `OnRealtimeDisconnect` - Triggered when realtime clients disconnect
- `OnRealtimeSubscribe` - Triggered when realtime subscriptions are created
- `OnRealtimeMessage` - Triggered when realtime messages are sent

### 3. Creating Custom Hook Handlers

To create a new hook handler:

1. Add your handler function to the appropriate file in `internal/handlers/hook/`
2. Register the handler in `internal/hooks/hooks.go`

Example:

```go
// In internal/handlers/hook/record_hooks.go
func HandleCustomRecordEvent(e *core.RecordEvent) error {
    // Your custom logic here
    e.App.Logger().Info("Custom record event", "id", e.Record.Id)
    
    // Always call e.Next() to continue the execution chain
    return e.Next()
}

// In internal/hooks/hooks.go
func registerRecordHooks(app *pocketbase.PocketBase) {
    // Add your custom hook registration
    app.OnRecordCreate("specific_collection").BindFunc(func(e *core.RecordEvent) error {
        return hook.HandleCustomRecordEvent(e)
    })
}
```

### 4. Collection-Specific Hooks

You can register hooks for specific collections:

```go
// Only trigger for "users" collection
app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
    return hook.HandleUserCreate(e)
})

// Trigger for multiple collections
app.OnRecordCreate("users", "profiles").BindFunc(func(e *core.RecordEvent) error {
    return hook.HandleUserOrProfileCreate(e)
})
```

### 5. Hook Execution Order

Hooks are executed in the order they are registered. You can control execution order using the `Bind` method with priority:

```go
app.OnRecordCreate().Bind(&hook.Handler[*core.RecordEvent]{
    Id:       "high_priority_handler",
    Priority: 100, // Higher numbers execute first
    Func: func(e *core.RecordEvent) error {
        return hook.HandleHighPriorityEvent(e)
    },
})
```

## Best Practices

1. **Always call `e.Next()`** - This continues the execution chain
2. **Handle errors gracefully** - Return errors to stop execution
3. **Use appropriate log levels** - Debug for verbose, Info for important events
4. **Avoid blocking operations** - Keep hook handlers fast
5. **Use collection-specific hooks** when possible for better performance
6. **Document your custom hooks** - Add comments explaining the purpose

## Common Use Cases

- **Audit logging** - Track all record changes
- **Data validation** - Additional validation beyond schema
- **Notifications** - Send emails or push notifications on events
- **Data synchronization** - Update related records or external systems
- **Access control** - Additional permission checks
- **Analytics** - Track user behavior and system usage
- **Caching** - Invalidate caches when data changes

## Error Handling

If a hook handler returns an error, it will stop the execution chain and the original operation will fail. Use this for validation or to prevent unwanted operations:

```go
func HandleRecordCreate(e *core.RecordEvent) error {
    // Validation logic
    if someCondition {
        return errors.New("validation failed")
    }
    
    return e.Next()
}
```

## Testing Hooks

When testing, you can temporarily disable hooks or add test-specific hooks:

```go
// In tests, you might want to unbind certain hooks
app.OnRecordCreate().UnbindAll()

// Or add test-specific hooks
app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
    // Test-specific logic
    return e.Next()
})
```