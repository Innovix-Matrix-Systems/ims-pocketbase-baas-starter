package common

var ExcludedPaths = []string{
	"/api/collections/users/auth-with-password",
	"/api/collections/users/auth-refresh",
	"/api/collections/users/request-password-reset",
	"/api/collections/users/confirm-password-reset",
	"/api/collections/users/request-verification",
	"/api/collections/users/confirm-verification",
	"/api/collections/users/request-email-change",
	"/api/collections/users/confirm-email-change",
	// Exclude PocketBase system URLs
	"/api/health",
	"/api/settings",
	"/api/logs",
	"/api/files",
	// Exclude metrics endpoint
	"/metrics",
	// Exclude superuser collection auth endpoints
	"/api/collections/_superusers/auth-with-password",
	"/api/collections/_superusers/auth-refresh",
}
