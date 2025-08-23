# Custom Routes Guide

This document explains how to create custom API routes and endpoints in the IMS PocketBase BaaS Starter, including integration with API documentation.

## Overview

While PocketBase provides automatic CRUD APIs for collections, you often need custom endpoints for specific business logic, data processing, or integrations. This guide shows how to create, organize, and document custom routes.

## Project Structure for Custom Routes

```
internal/
├── handlers/
│   └── route/              # Custom route handlers
│       ├── user_handler.go
│       ├── stats_handler.go
│       └── cache_handler.go
├── routes/
│   └── routes.go          # Route registration
└── app/
    └── app.go             # App initialization with route registration
```

## Creating Your First Custom Route

### Step 1: Create a Route Handler

Create a new handler file in `internal/handlers/route/`:

```go
// internal/handlers/route/hello_handler.go
package route

import (
    "ims-pocketbase-baas-starter/pkg/logger"
    "github.com/pocketbase/pocketbase/core"
)

// HelloResponse represents the response structure
type HelloResponse struct {
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
    Version   string `json:"version"`
}

// HandleHello handles GET /api/v1/hello
func HandleHello(e *core.RequestEvent) error {
    logger := logger.GetLogger(e.App)
    logger.Info("Hello endpoint called")

    response := HelloResponse{
        Message:   "Hello from IMS PocketBase BaaS Starter!",
        Timestamp: time.Now().Format(time.RFC3339),
        Version:   "1.0.0",
    }

    return e.JSON(200, response)
}

// HandleHelloWithName handles GET /api/v1/hello/{name}
func HandleHelloWithName(e *core.RequestEvent) error {
    name := e.Request.PathValue("name")

    if name == "" {
        return e.JSON(400, map[string]string{
            "error": "Name parameter is required",
        })
    }

    response := map[string]interface{}{
        "message": fmt.Sprintf("Hello, %s!", name),
        "name":    name,
        "timestamp": time.Now().Format(time.RFC3339),
    }

    return e.JSON(200, response)
}
```

### Step 2: Register the Route

Add your route to `internal/routes/routes.go` following the consistent pattern:

```go
// internal/routes/routes.go
package routes

import (
    "ims-pocketbase-baas-starter/internal/handlers/route"
    "ims-pocketbase-baas-starter/internal/middlewares"
    "ims-pocketbase-baas-starter/pkg/logger"

    "github.com/pocketbase/pocketbase/core"
)

// Route represents a custom application route with its configuration
type Route struct {
    Method      string                           // HTTP method (GET, POST, PUT, DELETE, etc.)
    Path        string                           // Route path
    Handler     func(*core.RequestEvent) error   // Handler function to execute when route is called
    Middlewares []func(*core.RequestEvent) error // Middlewares to apply to this route
    Enabled     bool                             // Whether the route should be registered
    Description string                           // Human-readable description of what the route does
}

// RegisterCustom registers all custom routes with the PocketBase application
func RegisterCustom(e *core.ServeEvent) {
    authMiddleware := middlewares.NewAuthMiddleware()
    logger := logger.GetLogger(e.App)

    g := e.Router.Group("/api/v1")

    // Define all custom routes in a consistent array structure
    routes := []Route{
        {
            Method:      "GET",
            Path:        "/hello",
            Handler:     route.HandleHello,
            Middlewares: []func(*core.RequestEvent) error{},
            Enabled:     true,
            Description: "Public hello world route",
        },
        {
            Method:  "GET",
            Path:    "/hello/{name}",
            Handler: route.HandleHelloWithName,
            Middlewares: []func(*core.RequestEvent) error{
                authMiddleware.RequireAuthFunc(), // Apply authentication middleware
            },
            Enabled:     true,
            Description: "Personalized hello route (auth required)",
        },
    }

    // Register enabled routes
    for _, route := range routes {
        if !route.Enabled {
            continue
        }

        // Create the final handler with middlewares applied
        finalHandler := route.Handler
        for i := len(route.Middlewares) - 1; i >= 0; i-- {
            middleware := route.Middlewares[i]
            nextHandler := finalHandler
            finalHandler = func(e *core.RequestEvent) error {
                if err := middleware(e); err != nil {
                    return err
                }
                return nextHandler(e)
            }
        }

        // Register the route with the appropriate HTTP method
        switch route.Method {
        case "GET":
            g.GET(route.Path, finalHandler)
        case "POST":
            g.POST(route.Path, finalHandler)
        case "PUT":
            g.PUT(route.Path, finalHandler)
        case "DELETE":
            g.DELETE(route.Path, finalHandler)
        case "PATCH":
            g.PATCH(route.Path, finalHandler)
        }
    }

    logger.Info("Custom routes registered successfully")
}
```

### Step 3: Register Routes in App

Ensure your routes are registered in `internal/app/app.go`:

```go
// internal/app/app.go
func NewApp() *pocketbase.PocketBase {
    app := pocketbase.New()

    // ... other initialization code ...

    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        // Register custom routes
        routes.RegisterCustom(se)

        return se.Next()
    })

    return app
}
```

## Advanced Route Patterns

### 1. Routes with Authentication

Create protected routes that require authentication:

```go
// internal/handlers/route/protected_handler.go
func HandleProtectedEndpoint(e *core.RequestEvent) error {
    // Get the authenticated user
    authRecord := e.Auth
    if authRecord == nil {
        return e.JSON(401, map[string]string{
            "error": "Authentication required",
        })
    }

    response := map[string]interface{}{
        "message": "This is a protected endpoint",
        "user_id": authRecord.Id,
        "user_email": authRecord.Email(),
    }

    return e.JSON(200, response)
}
```

Register with authentication middleware:

```go
// In routes.go
func RegisterCustomRoutes(e *core.ServeEvent) error {
    // Protected routes group
    e.Router.GET("/api/v1/protected/profile", route.HandleProtectedEndpoint,
        middlewares.RequireAuth())

    return nil
}
```

### 2. Routes with Request Validation

Handle POST requests with validation:

```go
// internal/handlers/route/user_handler.go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

type CreateUserResponse struct {
    ID      string `json:"id"`
    Message string `json:"message"`
}

func HandleCreateUser(e *core.RequestEvent) error {
    var req CreateUserRequest

    // Parse JSON body
    if err := e.BindBody(&req); err != nil {
        return e.JSON(400, map[string]string{
            "error": "Invalid JSON format",
        })
    }

    // Validate request
    if err := validateRequest(req); err != nil {
        return e.JSON(400, map[string]string{
            "error": err.Error(),
        })
    }

    // Create user logic here
    userID, err := createUser(e.App, req)
    if err != nil {
        return e.JSON(500, map[string]string{
            "error": "Failed to create user",
        })
    }

    response := CreateUserResponse{
        ID:      userID,
        Message: "User created successfully",
    }

    return e.JSON(201, response)
}

func validateRequest(req CreateUserRequest) error {
    if req.Name == "" {
        return errors.New("name is required")
    }
    if req.Email == "" {
        return errors.New("email is required")
    }
    if req.Age < 18 {
        return errors.New("age must be at least 18")
    }
    return nil
}
```

### 3. File Upload Routes

Handle file uploads:

```go
// internal/handlers/route/upload_handler.go
func HandleFileUpload(e *core.RequestEvent) error {
    // Parse multipart form
    err := e.Request.ParseMultipartForm(10 << 20) // 10 MB max
    if err != nil {
        return e.JSON(400, map[string]string{
            "error": "Failed to parse multipart form",
        })
    }

    file, header, err := e.Request.FormFile("file")
    if err != nil {
        return e.JSON(400, map[string]string{
            "error": "No file provided",
        })
    }
    defer file.Close()

    // Validate file type
    if !isValidFileType(header.Header.Get("Content-Type")) {
        return e.JSON(400, map[string]string{
            "error": "Invalid file type",
        })
    }

    // Save file logic here
    fileID, err := saveFile(e.App, file, header)
    if err != nil {
        return e.JSON(500, map[string]string{
            "error": "Failed to save file",
        })
    }

    return e.JSON(200, map[string]interface{}{
        "file_id": fileID,
        "filename": header.Filename,
        "size": header.Size,
    })
}
```

## Integration with API Documentation

### Step 1: Register Custom Routes in API Docs

Add your custom routes to the API Docs generator:

```go
// internal/app/app.go (in the OnServe handler)
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Initialize API docs generator
    generator := apidoc.InitializeGenerator(se.App)

    // Register custom routes for documentation
    registerApiDocRoutes(generator)

    return se.Next()
})

func registerApiDocRoutes(generator *apidoc.Generator) {
    // Hello endpoint
    generator.AddCustomRoute(apidoc.CustomRoute{
        Method:      "GET",
        Path:        "/api/v1/hello",
        Summary:     "Hello World",
        Description: "Returns a simple greeting message",
        Tags:        []string{"General"},
        Protected:   false,
    })

    // Hello with name endpoint
    generator.AddCustomRoute(apidoc.CustomRoute{
        Method:      "GET",
        Path:        "/api/v1/hello/{name}",
        Summary:     "Personalized Hello",
        Description: "Returns a personalized greeting message",
        Tags:        []string{"General"},
        Protected:   false,
        Parameters: []apidoc.Parameter{
            {
                Name:        "name",
                In:          "path",
                Required:    true,
                Description: "Name for personalized greeting",
                Schema: map[string]interface{}{
                    "type": "string",
                },
            },
        },
    })

    // Protected endpoint
    generator.AddCustomRoute(apidoc.CustomRoute{
        Method:      "GET",
        Path:        "/api/v1/protected/profile",
        Summary:     "Get User Profile",
        Description: "Returns authenticated user's profile information",
        Tags:        []string{"User", "Protected"},
        Protected:   true, // Adds authentication requirement
    })

    // Create user endpoint
    generator.AddCustomRoute(apidoc.CustomRoute{
        Method:      "POST",
        Path:        "/api/v1/users",
        Summary:     "Create User",
        Description: "Creates a new user with validation",
        Tags:        []string{"User"},
        Protected:   false,
        RequestBody: &apidoc.RequestBody{
            Required:    true,
            Description: "User creation data",
            Content: map[string]apidoc.MediaType{
                "application/json": {
                    Schema: map[string]interface{}{
                        "type": "object",
                        "required": []string{"name", "email", "age"},
                        "properties": map[string]interface{}{
                            "name": map[string]interface{}{
                                "type":      "string",
                                "minLength": 2,
                                "maxLength": 50,
                            },
                            "email": map[string]interface{}{
                                "type":   "string",
                                "format": "email",
                            },
                            "age": map[string]interface{}{
                                "type":    "integer",
                                "minimum": 18,
                                "maximum": 120,
                            },
                        },
                    },
                },
            },
        },
    })

    // File upload endpoint
    generator.AddCustomRoute(apidoc.CustomRoute{
        Method:      "POST",
        Path:        "/api/v1/upload",
        Summary:     "Upload File",
        Description: "Uploads a file to the server",
        Tags:        []string{"Files"},
        Protected:   true,
        RequestBody: &apidoc.RequestBody{
            Required:    true,
            Description: "File to upload",
            Content: map[string]apidoc.MediaType{
                "multipart/form-data": {
                    Schema: map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "file": map[string]interface{}{
                                "type":   "string",
                                "format": "binary",
                            },
                        },
                    },
                },
            },
        },
    })
}
```

### Step 2: Access Your Documentation

After registering your routes, they will appear in:

- **API Docs**: http://localhost:8090/api-docs
- **ReDoc**: http://localhost:8090/api-docs/redoc
- **OpenAPI JSON**: http://localhost:8090/api-docs/openapi.json

## Route Organization Best Practices

### 1. Group Related Routes

```go
// internal/handlers/route/user_routes.go
func RegisterUserRoutes(router *echo.Group) {
    userGroup := router.Group("/users")

    userGroup.GET("", route.HandleListUsers)
    userGroup.POST("", route.HandleCreateUser)
    userGroup.GET("/:id", route.HandleGetUser)
    userGroup.PUT("/:id", route.HandleUpdateUser)
    userGroup.DELETE("/:id", route.HandleDeleteUser)
}

// internal/handlers/route/admin_routes.go
func RegisterAdminRoutes(router *echo.Group) {
    adminGroup := router.Group("/admin", middlewares.RequireAdmin())

    adminGroup.GET("/stats", route.HandleAdminStats)
    adminGroup.GET("/users", route.HandleAdminListUsers)
    adminGroup.POST("/maintenance", route.HandleMaintenanceMode)
}
```

### 2. Use Consistent Response Formats

```go
// pkg/common/response.go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    PerPage    int `json:"per_page,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
    Total      int `json:"total,omitempty"`
}

func SuccessResponse(data interface{}) APIResponse {
    return APIResponse{
        Success: true,
        Data:    data,
    }
}

func ErrorResponse(message string) APIResponse {
    return APIResponse{
        Success: false,
        Error:   message,
    }
}
```

### 3. Implement Proper Error Handling

```go
// internal/handlers/route/error_handler.go
func HandleWithErrorRecovery(handler func(*core.RequestEvent) error) func(*core.RequestEvent) error {
    return func(e *core.RequestEvent) error {
        defer func() {
            if r := recover(); r != nil {
                logger := logger.GetLogger(e.App)
                logger.Error("Route handler panic", "error", r)

                e.JSON(500, ErrorResponse("Internal server error"))
            }
        }()

        return handler(e)
    }
}

// Usage
e.Router.GET("/api/v1/risky-endpoint", HandleWithErrorRecovery(route.HandleRiskyOperation))
```

## Testing Custom Routes

### Unit Testing Route Handlers

```go
// internal/handlers/route/hello_handler_test.go
func TestHandleHello(t *testing.T) {
    // Create test app
    app := pocketbase.NewWithConfig(pocketbase.Config{
        DefaultDebug: false,
    })

    // Create test request
    req := httptest.NewRequest("GET", "/api/v1/hello", nil)
    rec := httptest.NewRecorder()

    // Create request event
    e := &core.RequestEvent{
        App:      app,
        Request:  req,
        Response: rec,
    }

    // Call handler
    err := route.HandleHello(e)

    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 200, rec.Code)

    var response HelloResponse
    err = json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "Hello from IMS PocketBase BaaS Starter!", response.Message)
}
```

### Integration Testing

```go
// tests/integration/routes_test.go
func TestCustomRoutesIntegration(t *testing.T) {
    // Start test server
    app := setupTestApp()
    server := httptest.NewServer(app.Router())
    defer server.Close()

    // Test hello endpoint
    resp, err := http.Get(server.URL + "/api/v1/hello")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    // Test protected endpoint without auth
    resp, err = http.Get(server.URL + "/api/v1/protected/profile")
    assert.NoError(t, err)
    assert.Equal(t, 401, resp.StatusCode)
}
```

## Performance Considerations

### 1. Use Caching for Expensive Operations

```go
func HandleExpensiveStats(e *core.RequestEvent) error {
    cacheService := cache.GetInstance()
    cacheKey := "expensive_stats"

    // Try cache first
    if cachedData, found := cacheService.Get(cacheKey); found {
        return e.JSON(200, cachedData)
    }

    // Compute expensive stats
    stats, err := computeExpensiveStats(e.App)
    if err != nil {
        return e.JSON(500, ErrorResponse("Failed to compute stats"))
    }

    // Cache for 5 minutes
    cacheService.Set(cacheKey, stats, 5*time.Minute)

    return e.JSON(200, SuccessResponse(stats))
}
```

### 2. Implement Request Timeouts

```go
func HandleLongRunningOperation(e *core.RequestEvent) error {
    ctx, cancel := context.WithTimeout(e.Request.Context(), 30*time.Second)
    defer cancel()

    result, err := performLongOperation(ctx)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return e.JSON(408, ErrorResponse("Request timeout"))
        }
        return e.JSON(500, ErrorResponse("Operation failed"))
    }

    return e.JSON(200, SuccessResponse(result))
}
```

## Security Best Practices

### 1. Input Validation

```go
func validateAndSanitizeInput(input string) (string, error) {
    // Remove dangerous characters
    sanitized := strings.TrimSpace(input)

    // Validate length
    if len(sanitized) == 0 {
        return "", errors.New("input cannot be empty")
    }

    if len(sanitized) > 1000 {
        return "", errors.New("input too long")
    }

    return sanitized, nil
}
```

### 2. Rate Limiting

```go
func rateLimitMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Implement rate limiting logic
        clientIP := c.RealIP()

        // Check rate limit using cache
        cacheService := cache.GetInstance()
        key := fmt.Sprintf("rate_limit:%s", clientIP)

        if _, found := cacheService.Get(key); found {
            return c.JSON(429, ErrorResponse("Rate limit exceeded"))
        }

        // Set rate limit
        cacheService.Set(key, true, 1*time.Minute)

        return next(c)
    }
}
```

This guide provides a comprehensive foundation for creating, organizing, and documenting custom routes in your PocketBase application.
