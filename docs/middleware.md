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

The middleware is located in `internal/middlewares/auth.go` and provides:

- `AuthMiddleware` struct for organizing middleware functionality
- `NewAuthMiddleware()` constructor function
- `RequireAuth()` method that returns a PocketBase hook handler
- `RequireAuthFunc()` method that returns a middleware function
- `BindToRouter()` method for future router integration

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
