package constant

import domainerr "sipon-api/internal/domain/errors"

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type RoleName string

type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeCustom RoleType = "custom"
)

type ScopeType string

const (
	ScopeTypeGlobal    ScopeType = "global"
	ScopeTypeRegion    ScopeType = "region"
	ScopeTypeCommunity ScopeType = "community"
)

// RoleInit adalah tipe pendaftaran system role — analog dengan PermissionInit
// untuk permission. Key-nya RoleName, dan itu SATU-SATUNYA tempat mendaftarkan
// system role beserta metadatanya; RoleSeeder (internal/seeders/role_seeder.go)
// membaca langsung dari DefaultRolesInit, jadi tidak ada daftar role kedua yang
// bisa tidak sinkron.
type RoleInit map[RoleName]struct {
	RoleType    RoleType
	ScopeType   ScopeType
	DisplayName string
	Description string
	Assignable  bool
}

const (
	UserGodRoleName    RoleName = "usergod"
	SuperAdminRoleName RoleName = "superadmin"
	AdminRoleName      RoleName = "admin"
	MemberRoleName     RoleName = "member"
)

// DefaultRolesInit adalah SATU-SATUNYA tempat mendaftarkan system role.
// Menambah system role baru = menambah constant RoleName di atas + entry di sini
// — RoleSeeder akan otomatis meng-upsert-nya ke tabel roles.
var DefaultRolesInit = RoleInit{
	UserGodRoleName: {
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		DisplayName: "Developer / Vendor",
		Description: "Akses penuh ke seluruh sistem. Hanya untuk developer — tidak bisa di-assign via UI.",
		Assignable:  false,
	},
	SuperAdminRoleName: {
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		DisplayName: "Super Admin",
		Description: "Admin pusat, mengelola seluruh platform.",
		Assignable:  true,
	},
	AdminRoleName: {
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		DisplayName: "Admin",
		Description: "Administrator sistem dengan akses pengelolaan global.",
		Assignable:  true,
	},
	MemberRoleName: {
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		DisplayName: "Member",
		Description: "User terdaftar biasa.",
		Assignable:  true,
	},
}

const (
	CodeRoleIDRequired              domainerr.Code = "DOMAIN_ROLE_ID_REQUIRED"
	CodeRoleNameRequired            domainerr.Code = "DOMAIN_ROLE_NAME_REQUIRED"
	CodeRoleDisplayNameRequired     domainerr.Code = "DOMAIN_ROLE_DISPLAY_NAME_REQUIRED"
	CodeRoleTypeInvalid             domainerr.Code = "DOMAIN_ROLE_TYPE_INVALID"
	CodeRoleScopeTypeInvalid        domainerr.Code = "DOMAIN_ROLE_SCOPE_TYPE_INVALID"
	CodeRoleNotFound                domainerr.Code = "DOMAIN_ROLE_NOT_FOUND"
	CodeRoleQueryFailed             domainerr.Code = "DOMAIN_ROLE_QUERY_FAILED"
	CodeRolePersistenceFailed       domainerr.Code = "DOMAIN_ROLE_PERSISTENCE_FAILED"
	CodeRoleDuplicateName           domainerr.Code = "DOMAIN_ROLE_DUPLICATE_NAME"
	CodeRoleNotAssignable           domainerr.Code = "DOMAIN_ROLE_NOT_ASSIGNABLE"
	CodeRoleIsSystem                domainerr.Code = "DOMAIN_ROLE_IS_SYSTEM"
	CodeUserRoleIDRequired          domainerr.Code = "DOMAIN_USER_ROLE_ID_REQUIRED"
	CodeUserRoleUserIDRequired      domainerr.Code = "DOMAIN_USER_ROLE_USER_ID_REQUIRED"
	CodeUserRoleRoleIDRequired      domainerr.Code = "DOMAIN_USER_ROLE_ROLE_ID_REQUIRED"
	CodeUserRoleScopeTypeInvalid    domainerr.Code = "DOMAIN_USER_ROLE_SCOPE_TYPE_INVALID"
	CodeUserRoleScopeIDRequired     domainerr.Code = "DOMAIN_USER_ROLE_SCOPE_ID_REQUIRED"
	CodeUserRoleScopeIDMustBeEmpty  domainerr.Code = "DOMAIN_USER_ROLE_SCOPE_ID_MUST_BE_EMPTY"
	CodeUserRoleNotFound            domainerr.Code = "DOMAIN_USER_ROLE_NOT_FOUND"
	CodeUserRoleQueryFailed         domainerr.Code = "DOMAIN_USER_ROLE_QUERY_FAILED"
	CodeUserRolePersistenceFailed   domainerr.Code = "DOMAIN_USER_ROLE_PERSISTENCE_FAILED"
	CodeUserRoleDuplicate           domainerr.Code = "DOMAIN_USER_ROLE_DUPLICATE"
	CodeRoleUserIDRequired          domainerr.Code = "DOMAIN_ROLE_USER_ID_REQUIRED"
	CodeRoleUserAssignmentDuplicate domainerr.Code = "DOMAIN_ROLE_USER_ASSIGNMENT_DUPLICATE"
	CodeRoleUserAssignmentNotFound  domainerr.Code = "DOMAIN_ROLE_USER_ASSIGNMENT_NOT_FOUND"
	CodeRoleAssignmentScopeMismatch domainerr.Code = "DOMAIN_ROLE_ASSIGNMENT_SCOPE_MISMATCH"

	CodeRolePermissionKeyRequired       domainerr.Code = "DOMAIN_ROLE_PERMISSION_KEY_REQUIRED"
	CodeRolePermissionKeyInvalid        domainerr.Code = "DOMAIN_ROLE_PERMISSION_KEY_INVALID"
	CodeRolePermissionRequiresCustom    domainerr.Code = "DOMAIN_ROLE_PERMISSION_REQUIRES_CUSTOM_ROLE"
	CodeRolePermissionNotFound          domainerr.Code = "DOMAIN_ROLE_PERMISSION_NOT_FOUND"
	CodeRolePermissionDuplicate         domainerr.Code = "DOMAIN_ROLE_PERMISSION_DUPLICATE"
	CodeRolePermissionQueryFailed       domainerr.Code = "DOMAIN_ROLE_PERMISSION_QUERY_FAILED"
	CodeRolePermissionPersistenceFailed domainerr.Code = "DOMAIN_ROLE_PERMISSION_PERSISTENCE_FAILED"
)
