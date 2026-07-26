package principal

import "strings"

type Role struct {
	Name      string
	RoleType  string // system | custom
	ScopeType string // global | region | community
	ScopeID   *string
}

type Permission struct {
	Key   string
	Scope string
}

type Principal struct {
	UserID      string
	SessionID   string
	Roles       []Role
	Permissions []Permission
}

func (p *Principal) HasRole(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, r := range p.Roles {
		if strings.ToLower(r.Name) == name {
			return true
		}
	}
	return false
}

func (p *Principal) HasPermission(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, perm := range p.Permissions {
		if strings.ToLower(perm.Key) == key {
			return true
		}
	}
	return false
}

func (p *Principal) IsUsergod() bool {
	return p.HasRole("usergod")
}

func (p *Principal) IsSuperAdmin() bool {
	return p.HasRole("superadmin") || p.IsUsergod()
}

func (p *Principal) IsAdmin() bool {
	return p.HasRole("admin") || p.IsSuperAdmin()
}
