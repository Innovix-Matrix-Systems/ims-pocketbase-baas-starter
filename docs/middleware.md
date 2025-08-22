# Custom Middleware Setup Guide

This guide explains how to set up and use custom authentication middleware in your PocketBase application.

## Overview

The authentication middleware provides a clean interface for protecting routes using PocketBase's built-in `apis.RequireAuth()` functionality. It can be applied to both custom routes and default PocketBase API endpoints.

### PocketBase Built-in Middleware APIs

PocketBase provides several built-in middleware helpers that you can use in your Go application:

- `apis.RequireAuth(...)` - Requires authentication from any or specific auth collections
- `apis.RequireGuestOnly()` - Allows only unauthenticated requests
- `apis.RequireSuperuserAuth()` - Requires superuser/admin authentication
- `apis.RequireSuperuserOrOwnerAuth(...)` - Requires superuser or record owner authentication
- `apis.BodyLimit(...)` - Limits request body size
- `apis.Gzip()` - Enables gzip compression
- And more...

For complete documentation on PocketBase routing and middleware, see: [Extend with Go - Routing - PocketBase Docs](https://pocketbase.io/docs/go-routing/)

## Middleware Structure

The middleware system is located in `internal/middlewares/` and follows a consistent pattern similar to routes and cron jobs:

- `middlewares.go` - Main middleware registration following the same pattern as routes/crons
- `auth.go` - Authentication middleware implementation
- `metrics.go` - Metrics collection middleware implementation
- `permission.go` - Permission-based access control middleware implementation

### Middleware Registration Pattern

The middleware system uses a consistent array-based structure:

```go
// Middleware represents an application middleware with its configuration
type Middleware struct {
    ID          string                         // Unique identifier for the middleware
    Handler     func(*core.RequestEvent) error // Handler function to execute
    Enabled     bool                           // Whether the middleware should be registered
    Description string                         // Human-readable description of what the middleware does
    Order       int                            // Order of execution (lower numbers execute first)
}

// RegisterMiddlewares registers all application middlewares with the PocketBase router
func RegisterMiddlewares(e *core.ServeEvent) {
    // Define all middlewares in a consistent array structure
    middlewares := []Middleware{
        {
            ID:          "metricsCollection",
            Handler:     getMetricsMiddlewareHandler(),
            Enabled:     true,
            Description: "Collect HTTP request metrics",
            Order:       1,
        },
        {
            ID:          "jwtAuth",
            Handler:     getAuthMiddlewareHandler(e),
            Enabled:     true,
            Description: "JWT authentication with exclusions",
            Order:       2,
        },
    }

    // Register enabled middlewares
    for _, middleware := range middlewares {
        if !middleware.Enabled {
            continue
        }

        e.Router.Bind(&hook.Handler[*core.RequestEvent]{
            Id:   middleware.ID,
            Func: middleware.Handler,
        })
    }
}
```

## Basic Usage

### 1. Initialize the Middleware

```go
middleware := middlewares.NewAuthMiddleware()
```

### 2. Get Authentication Function

```go
// Any authenticated user from any auth collection
authFunc := middleware.RequireAuthFunc()

// Only users from specific collections
authFunc := middleware.RequireAuthFunc("users")
authFunc := middleware.RequireAuthFunc("users", "admins")
```

## Protecting Custom Routes

### Method 1: Apply Middleware Inside Handler (Recommended)

```go
func RegisterCustom(e *core.ServeEvent) {
    middleware := middlewares.NewAuthMiddleware()

    g := e.Router.Group("/api/v1")

    // Public route (no auth required)
    g.GET("/hello", func(e *core.RequestEvent) error {
        return e.JSON(200, map[string]string{"msg": "Hello from custom route"})
    })

    // Protected route (auth required)
    g.GET("/protected", func(e *core.RequestEvent) error {
        // Apply authentication middleware
        authFunc := middleware.RequireAuthFunc()
        if err := authFunc(e); err != nil {
            return err
        }

        // Your protected handler logic
        return e.JSON(200, map[string]string{"msg": "You are authenticated!"})
    })
}
```

## Protecting Default PocketBase Routes

### Method 1: Using OnRecordRequest Hooks (Recommended)

Apply authentication to all record operations:

```go
func Run() {
    app := pocketbase.New()

    // Initialize middleware
    middleware := middlewares.NewAuthMiddleware()

    // Apply auth to all record operations
    app.OnRecordRequest().BindFunc(func(e *core.RecordRequestEvent) error {
        authFunc := middleware.RequireAuthFunc()
        if err := authFunc(e.RequestEvent); err != nil {
            return err
        }
        return e.Next()
    })

    // Rest of your app setup...
}
```

### Method 2: Using OnServe Hook for Specific Routes

Apply authentication to specific PocketBase API endpoints:

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    middleware := middlewares.NewAuthMiddleware()

    // Apply auth to specific PocketBase API endpoints
    se.Router.Bind(&hook.Handler[*core.RequestEvent]{
        Id: "customAuth",
        Func: func(e *core.RequestEvent) error {
            path := e.Request.URL.Path
            if strings.HasPrefix(path, "/api/collections/") {
                authFunc := middleware.RequireAuthFunc()
                if err := authFunc(e); err != nil {
                    return err
                }
            }
            return e.Next()
        },
    })

    // Your existing routes...
    return se.Next()
})
```

### Method 3: Collection-Specific Hooks

Protect specific collection operations:

```go
// Protect specific collection operations
app.OnRecordListRequest("users").BindFunc(func(e *core.RecordListRequestEvent) error {
    authFunc := middleware.RequireAuthFunc()
    if err := authFunc(e.RequestEvent); err != nil {
        return err
    }
    return e.Next()
})

app.OnRecordViewRequest("users").BindFunc(func(e *core.RecordViewRequestEvent) error {
    authFunc := middleware.RequireAuthFunc()
    if err := authFunc(e.RequestEvent); err != nil {
        return err
    }
    return e.Next()
})
```

## Collection Filtering

You can restrict authentication to specific collections:

```go
// Any auth collection (default)
authFunc := middleware.RequireAuthFunc()

// Only "users" collection
authFunc := middleware.RequireAuthFunc("users")

// Multiple collections
authFunc := middleware.RequireAuthFunc("users", "admins")
```

## Error Handling

The middleware uses PocketBase's standard error handling:

- **401 Unauthorized**: Returned when authentication fails
- **Standard Format**: Follows PocketBase's error response format

Example error response:

```json
{
  "code": 401,
  "message": "The request requires valid record authorization token to be set.",
  "data": {}
}
```

## Testing Your Middleware

### 1. Test Public Routes

```bash
curl http://localhost:8090/api/v1/hello
# Should return: {"msg": "Hello from custom route"}
```

### 2. Test Protected Routes Without Auth

```bash
curl http://localhost:8090/api/v1/protected
# Should return: 401 Unauthorized
```

### 3. Test Protected Routes With Auth

```bash
# First, get an auth token
curl -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{"identity": "user@example.com", "password": "password"}'

# Use the token
curl http://localhost:8090/api/v1/protected \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
# Should return: {"msg": "You are authenticated!"}
```

## Important Considerations

### 1. Avoid Breaking Auth Endpoints

Be careful not to apply authentication to PocketBase's auth endpoints:

- `/api/collections/users/auth-with-password`
- `/api/collections/users/auth-refresh`
- `/api/collections/users/request-password-reset`

### 2. Collection Rules vs Middleware

Consider using PocketBase's built-in collection access rules for simpler use cases. Custom middleware is best for:

- Complex authentication logic
- Cross-collection validation
- Custom JWT validation (future enhancement)
- Logging and monitoring

### 3. Development vs Production

The middleware works the same in both environments, but consider:

- HTTPS enforcement in production
- Token storage security on the client side
- Rate limiting for auth endpoints

## Permission Middleware

The permission middleware extends the authentication system to provide permission-based access control for custom routes. It checks if an authenticated user has specific permissions before allowing access to protected resources.

### Overview

The permission middleware builds upon the existing RBAC (Role-Based Access Control) system where users can have:
- Direct permissions assigned to them
- Permissions inherited through roles

### Basic Setup

The permission middleware is located in `internal/middlewares/permission.go` and provides:

- `PermissionMiddleware` struct for organizing permission functionality
- `NewPermissionMiddleware()` constructor function
- `RequirePermission()` method that returns a middleware function
- `HasPermission()` method for checking user permissions

### Basic Usage with Single Permission

```go
func RegisterCustomRoutes(e *core.ServeEvent) {
    // Initialize middlewares
    authMiddleware := middlewares.NewAuthMiddleware()
    permMiddleware := middlewares.NewPermissionMiddleware()

    g := e.Router.Group("/api/v1")

    // Protected route requiring specific permission
    g.GET("/admin/users", func(e *core.RequestEvent) error {
        // First ensure user is authenticated
        authFunc := authMiddleware.RequireAuthFunc()
        if err := authFunc(e); err != nil {
            return err
        }

        // Then check for specific permission
        permFunc := permMiddleware.RequirePermission("users.view")
        if err := permFunc(e); err != nil {
            return err
        }

        // Handler logic for authorized users
        return e.JSON(200, map[string]string{"message": "User list access granted"})
    })
}
```

### Usage with Multiple Permissions (ANY Logic)

```go
// User needs ANY of these permissions to access the route
g.POST("/api/content", func(e *core.RequestEvent) error {
    // Authentication first
    authFunc := authMiddleware.RequireAuthFunc()
    if err := authFunc(e); err != nil {
        return err
    }

    // Permission check - user needs ANY of these permissions
    permFunc := permMiddleware.RequirePermission("content.create", "content.admin", "content.manage")
    if err := permFunc(e); err != nil {
        return err
    }

    // Handler logic
    return e.JSON(200, map[string]string{"message": "Content creation access granted"})
})
```

### Integration with Route Groups

```go
func RegisterAdminRoutes(e *core.ServeEvent) {
    authMiddleware := middlewares.NewAuthMiddleware()
    permMiddleware := middlewares.NewPermissionMiddleware()

    // Admin route group
    adminGroup := e.Router.Group("/api/admin")

    // Apply authentication and admin permission to all routes in the group
    adminGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Convert echo.Context to core.RequestEvent
            e := c.(*core.RequestEvent)
            
            // Apply authentication
            authFunc := authMiddleware.RequireAuthFunc()
            if err := authFunc(e); err != nil {
                return err
            }

            // Apply admin permission check
            permFunc := permMiddleware.RequirePermission("admin.access")
            if err := permFunc(e); err != nil {
                return err
            }

            return next(c)
        }
    })

    // All routes in this group now require authentication and admin.access permission
    adminGroup.GET("/dashboard", func(e *core.RequestEvent) error {
        return e.JSON(200, map[string]string{"message": "Admin dashboard"})
    })

    adminGroup.GET("/settings", func(e *core.RequestEvent) error {
        return e.JSON(200, map[string]string{"message": "Admin settings"})
    })
}
```

### Permission Error Handling

The permission middleware returns standard HTTP error responses:

- **403 Forbidden**: Returned when user is authenticated but lacks required permissions
- **401 Unauthorized**: Returned when user is not authenticated (handled by auth middleware)

Example error response for insufficient permissions:

```json
{
  "code": 403,
  "message": "You don't have permission to access this resource",
  "data": {}
}
```

### Testing Permission Middleware

#### 1. Test with User Having Required Permission

```bash
# First, authenticate and get a token for a user with the required permission
curl -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{"identity": "admin@example.com", "password": "password"}'

# Use the token to access protected resource
curl http://localhost:8090/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
# Should return: {"message": "User list access granted"}
```

#### 2. Test with User Lacking Required Permission

```bash
# Authenticate as a user without the required permission
curl -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{"identity": "user@example.com", "password": "password"}'

# Try to access protected resource
curl http://localhost:8090/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
# Should return: 403 Forbidden
```

### Permission System Requirements

For the permission middleware to work, your PocketBase application should have:

1. **Users Collection**: With authentication enabled
2. **Roles Collection**: For role-based permissions
3. **Permissions Collection**: Defining available permissions
4. **User-Role Relationships**: Users can have multiple roles
5. **Role-Permission Relationships**: Roles can have multiple permissions
6. **Direct User Permissions**: Users can have direct permissions without roles

### Best Practices

1. **Always Apply Authentication First**: Permission middleware should be used after authentication middleware
2. **Use Descriptive Permission Names**: Use clear, hierarchical permission names like `users.view`, `content.create`
3. **Group Related Routes**: Apply permissions at the route group level when possible
4. **Test Permission Scenarios**: Test with users having different permission combinations
5. **Admin Override**: Admin users typically bypass permission checks

## Future Extensions

The middleware is designed to be extensible. Future enhancements might include:

- Custom JWT validation logic
- Role-based authorization
- Token refresh handling
- Request logging and monitoring
- Rate limiting integration

## Troubleshooting

### Common Issues

1. **"too many arguments in call to g.GET"**

   - Solution: Apply middleware inside the handler, not as a separate parameter

2. **Auth not working on default routes**

   - Solution: Use OnRecordRequest hooks or OnServe hooks, not route-level middleware

3. **Can't login after applying middleware**
   - Solution: Exclude auth endpoints from middleware protection

### Debug Tips

1. Add logging to see which routes are being protected
2. Check the request path in middleware to ensure correct targeting
3. Verify auth tokens are being sent correctly in requests
4. Use PocketBase's admin UI to test authentication flows
