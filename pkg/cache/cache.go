package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheService provides a centralized caching solution for the application
type CacheService struct {
	cache *cache.Cache
	mu    sync.RWMutex
}

// CacheConfig holds configuration for the cache service
type CacheConfig struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

var (
	instance *CacheService
	once     sync.Once
)

// GetInstance returns the singleton cache service instance
func GetInstance() *CacheService {
	once.Do(func() {
		instance = NewCacheService(CacheConfig{
			DefaultExpiration: 10 * time.Minute, // Default 10 minutes
			CleanupInterval:   15 * time.Minute, // Cleanup every 15 minutes
		})
	})
	return instance
}

// NewCacheService creates a new cache service with the given configuration
func NewCacheService(config CacheConfig) *CacheService {
	return &CacheService{
		cache: cache.New(config.DefaultExpiration, config.CleanupInterval),
	}
}

// Set stores a value in the cache with the default expiration
func (cs *CacheService) Set(key string, value interface{}) {
	cs.cache.Set(key, value, cache.DefaultExpiration)
}

// SetWithExpiration stores a value in the cache with a custom expiration
func (cs *CacheService) SetWithExpiration(key string, value interface{}, expiration time.Duration) {
	cs.cache.Set(key, value, expiration)
}

// Get retrieves a value from the cache
func (cs *CacheService) Get(key string) (interface{}, bool) {
	return cs.cache.Get(key)
}

// GetString retrieves a string value from the cache
func (cs *CacheService) GetString(key string) (string, bool) {
	if value, found := cs.cache.Get(key); found {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetStringSlice retrieves a string slice from the cache
func (cs *CacheService) GetStringSlice(key string) ([]string, bool) {
	if value, found := cs.cache.Get(key); found {
		if slice, ok := value.([]string); ok {
			return slice, true
		}
	}
	return nil, false
}

// GetMap retrieves a map from the cache
func (cs *CacheService) GetMap(key string) (map[string]interface{}, bool) {
	if value, found := cs.cache.Get(key); found {
		if m, ok := value.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}

// Delete removes a value from the cache
func (cs *CacheService) Delete(key string) {
	cs.cache.Delete(key)
}

// DeletePattern removes all keys matching a pattern
func (cs *CacheService) DeletePattern(pattern string) int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	deleted := 0
	items := cs.cache.Items()

	for key := range items {
		// Simple pattern matching - you can enhance this with regex if needed
		if containsPattern(key, pattern) {
			cs.cache.Delete(key)
			deleted++
		}
	}

	return deleted
}

// Flush clears all items from the cache
func (cs *CacheService) Flush() {
	cs.cache.Flush()
}

// ItemCount returns the number of items in the cache
func (cs *CacheService) ItemCount() int {
	return cs.cache.ItemCount()
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats() map[string]interface{} {
	items := cs.cache.Items()

	stats := map[string]interface{}{
		"item_count": len(items),
		"items":      make(map[string]interface{}),
	}

	// Add item details (without actual values for security)
	itemDetails := make(map[string]interface{})
	for key, item := range items {
		itemDetails[key] = map[string]interface{}{
			"expires_at": item.Expiration,
			"expired":    item.Expired(),
		}
	}
	stats["items"] = itemDetails

	return stats
}

// InvalidateUserPermissions invalidates all user permission cache entries
func (cs *CacheService) InvalidateUserPermissions() int {
	return cs.DeletePattern("user_permissions_")
}

// InvalidateSwaggerCache invalidates all swagger-related cache entries
func (cs *CacheService) InvalidateSwaggerCache() int {
	return cs.DeletePattern("swagger_")
}

// InvalidateRoleCache invalidates all role-related cache entries
func (cs *CacheService) InvalidateRoleCache() int {
	return cs.DeletePattern("role_")
}

// InvalidatePermissionCache invalidates all permission-related cache entries
func (cs *CacheService) InvalidatePermissionCache() int {
	return cs.DeletePattern("permission_")
}

// Helper function for simple pattern matching
func containsPattern(key, pattern string) bool {
	// Simple contains check - you can enhance this with more sophisticated pattern matching
	return len(pattern) > 0 && len(key) >= len(pattern) && key[:len(pattern)] == pattern
}

// CacheKey generates standardized cache keys
type CacheKey struct{}

// UserPermissions generates a cache key for user permissions
func (CacheKey) UserPermissions(userID string) string {
	return fmt.Sprintf("user_permissions_%s", userID)
}

// RolePermissions generates a cache key for role permissions
func (CacheKey) RolePermissions(roleID string) string {
	return fmt.Sprintf("role_permissions_%s", roleID)
}

// RoleNames generates a cache key for role names batch
func (CacheKey) RoleNames(roleIDs []string) string {
	return fmt.Sprintf("role_names_%d_items", len(roleIDs))
}

// PermissionSlugs generates a cache key for permission slugs batch
func (CacheKey) PermissionSlugs(permissionIDs []string) string {
	return fmt.Sprintf("permission_slugs_%d_items", len(permissionIDs))
}

// SwaggerSpec generates a cache key for swagger specification
func (CacheKey) SwaggerSpec() string {
	return "swagger_spec"
}

// SwaggerCollectionsHash generates a cache key for collections hash
func (CacheKey) SwaggerCollectionsHash() string {
	return "swagger_collections_hash"
}

// BatchRoles generates a cache key for batch role data
func (CacheKey) BatchRoles() string {
	return "batch_roles_all"
}

// BatchPermissions generates a cache key for batch permission data
func (CacheKey) BatchPermissions() string {
	return "batch_permissions_all"
}
