package permission

// Permission constants for RBAC system
const (
	//system
	CacheClear = "cache.clear"
	// User permissions
	UserCreate           = "user.create"
	UserView             = "user.view"
	UserViewAll          = "user.view.all"
	UserUpdate           = "user.update"
	UserDelete           = "user.delete"
	UserRoleAssign       = "user.role.assign"
	UserPermissionAssign = "user.permission.assign"
	UserExport           = "user.export"

	// Role permissions
	RoleCreate  = "role.create"
	RoleView    = "role.view"
	RoleViewAll = "role.view.all"
	RoleUpdate  = "role.update"
	RoleDelete  = "role.delete"
)

// PermissionDefinition represents a permission with its metadata
type PermissionDefinition struct {
	Slug        string
	Name        string
	Description string
}

// GetAllPermissions returns all permission definitions
func GetAllPermissions() []PermissionDefinition {
	return []PermissionDefinition{
		{Slug: CacheClear, Name: "Clear Cache", Description: "Can clear the system cache"},
		{Slug: UserCreate, Name: "Create User", Description: "Can create new users"},
		{Slug: UserView, Name: "View User", Description: "Can view user details"},
		{Slug: UserViewAll, Name: "View All Users", Description: "Can view all users"},
		{Slug: UserUpdate, Name: "Update User", Description: "Can update user information"},
		{Slug: UserDelete, Name: "Delete User", Description: "Can delete users"},
		{Slug: UserRoleAssign, Name: "Assign Role To User", Description: "Can assign roles to users"},
		{Slug: UserPermissionAssign, Name: "Assign Permission To User", Description: "Can assign permissions to users"},
		{Slug: UserExport, Name: "Export Users", Description: "Can export user data as CSV"},
		{Slug: RoleCreate, Name: "Create Role", Description: "Can create new roles"},
		{Slug: RoleView, Name: "View Role", Description: "Can view role details"},
		{Slug: RoleViewAll, Name: "View All Roles", Description: "Can view all roles"},
		{Slug: RoleUpdate, Name: "Update Role", Description: "Can update role information"},
		{Slug: RoleDelete, Name: "Delete Role", Description: "Can delete roles"},
	}
}
