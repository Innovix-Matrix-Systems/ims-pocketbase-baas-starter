package internal

var ProtectedCollections = []string{"users", "roles", "permissions"}

var ExcludedPaths = []string{
	"/api/collections/users/auth-with-password",
	"/api/collections/users/auth-refresh",
	"/api/collections/users/request-password-reset",
	"/api/collections/users/confirm-password-reset",
	"/api/collections/users/request-verification",
	"/api/collections/users/confirm-verification",
	"/api/collections/users/request-email-change",
	"/api/collections/users/confirm-email-change",
}
