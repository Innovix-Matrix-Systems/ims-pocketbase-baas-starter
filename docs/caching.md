# Caching System Guide

This document explains how to use the built-in TTL (Time-To-Live) caching system in the IMS PocketBase BaaS Starter.

## Overview

The application includes a high-performance caching system with TTL support that can significantly improve performance for expensive operations like database queries or API calls.

> **Note:** This caching system is designed for **custom routes and endpoints** that you add to your application. PocketBase's built-in CRUD APIs are already optimized with SQLite performance and internal schema caching, so they don't need additional caching layers.

## Basic Usage

### Get Cache Instance

```go
// Get the cache instance (singleton pattern)
cacheService := cache.GetInstance()
```

### Store Data

```go
// Store data with TTL
cacheService.Set("user_profile_123", userData, 10*time.Minute)
cacheService.Set("session_token", token, 1*time.Hour)
cacheService.Set("temp_data", tempData, 30*time.Second)
```

### Retrieve Data

```go
// Get data from cache
if cachedData, found := cacheService.Get("user_profile_123"); found {
    userData := cachedData.(UserProfile) // Type assertion
    return userData
}

// Cache miss - fetch from source
userData := fetchFromDatabase()
```

### Clear Cache

```go
// Delete specific key
cacheService.Delete("user_profile_123")

// Delete multiple keys by pattern
cacheService.DeleteByPattern("user_profile:*")

// Clear all cache
cacheService.Flush()

// Get cache info
itemCount := cacheService.ItemCount()
```

## Cache Keys

Use structured, consistent cache key patterns:

```go
// Good patterns
"user:profile:123"
"search:golang:page:1"
"system:stats"
"session:abc123"

// Helper function
func buildUserKey(userID string) string {
    return fmt.Sprintf("user:profile:%s", userID)
}
```

## TTL Recommendations

```go
// Recommended TTL values
var cacheTTLs = map[string]time.Duration{
    "user_profile":   10 * time.Minute,  // User data
    "user_stats":     5 * time.Minute,   // Statistics
    "system_config":  30 * time.Minute,  // Configuration
    "search_results": 2 * time.Minute,   // Search results
    "static_content": 1 * time.Hour,     // Static data
}
```

## Cache Invalidation

### Automatic (TTL)

```go
// Cache expires automatically
cacheService.Set("temp_data", data, 5*time.Minute)
```

### Manual (Event-based)

```go
// In hook handlers (internal/handlers/hook/record_hooks.go)
func HandleUserUpdate(e *core.RecordEvent) error {
    cacheService := cache.GetInstance()

    // Clear specific user cache
    userKey := fmt.Sprintf("user:profile:%s", e.Record.Id)
    cacheService.Delete(userKey)

    // Clear related cache
    cacheService.Delete("user_stats")

    return e.Next()
}
```

## Common Pattern

```go
func getUserProfile(userID string) (*UserProfile, error) {
    cacheService := cache.GetInstance()
    cacheKey := fmt.Sprintf("user:profile:%s", userID)

    // Try cache first
    if cachedData, found := cacheService.Get(cacheKey); found {
        return cachedData.(*UserProfile), nil
    }

    // Cache miss - fetch from database
    profile, err := fetchUserFromDatabase(userID)
    if err != nil {
        return nil, err
    }

    // Cache the result
    cacheService.Set(cacheKey, profile, 10*time.Minute)

    return profile, nil
}
```

That's it! The cache system is simple and effective for improving your application's performance.
