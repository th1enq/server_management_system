package models

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type APIScope string

const (
	// Server management scopes
	ScopeServerRead   APIScope = "server:read"
	ScopeServerWrite  APIScope = "server:write"
	ScopeServerDelete APIScope = "server:delete"
	ScopeServerImport APIScope = "server:import"
	ScopeServerExport APIScope = "server:export"

	// User management scopes
	ScopeUserRead   APIScope = "user:read"
	ScopeUserWrite  APIScope = "user:write"
	ScopeUserDelete APIScope = "user:delete"

	// Report scopes
	ScopeReportRead  APIScope = "report:read"
	ScopeReportWrite APIScope = "report:write"

	// Profile scopes
	ScopeProfileRead  APIScope = "profile:read"
	ScopeProfileWrite APIScope = "profile:write"

	// Admin scopes
	ScopeAdminAll APIScope = "admin:all"
)

var AllScopes = []APIScope{
	ScopeServerRead, ScopeServerWrite, ScopeServerDelete, ScopeServerImport, ScopeServerExport,
	ScopeUserRead, ScopeUserWrite, ScopeUserDelete,
	ScopeReportRead, ScopeReportWrite,
	ScopeProfileRead, ScopeProfileWrite,
	ScopeAdminAll,
}

// GetDefaultScopes returns default scopes for a given role
func GetDefaultScopes(role UserRole) []APIScope {
	switch role {
	case RoleAdmin:
		return []APIScope{
			ScopeServerRead, ScopeServerWrite, ScopeServerDelete, ScopeServerImport, ScopeServerExport,
			ScopeUserRead, ScopeUserWrite, ScopeUserDelete,
			ScopeReportRead, ScopeReportWrite,
			ScopeProfileRead, ScopeProfileWrite,
			ScopeAdminAll,
		}
	case RoleUser:
		return []APIScope{
			ScopeServerRead, ScopeServerExport,
			ScopeReportRead,
			ScopeProfileRead, ScopeProfileWrite,
		}
	default:
		return []APIScope{ScopeProfileRead}
	}
}

func GetDefaultScopesMask(role UserRole) int64 {
	scopes := GetDefaultScopes(role)
	mask := int64(0)

	for i, scope := range AllScopes {
		for _, userScope := range scopes {
			if scope == userScope {
				mask |= 1 << i
				break
			}
		}
	}

	return mask
}

func ToBitmask(scopes []APIScope) int64 {
	mask := int64(0)

	for i, scope := range AllScopes {
		for _, userScope := range scopes {
			if scope == userScope {
				mask |= 1 << i
				break
			}
		}
	}

	return mask
}

func ToArray(mask int64) []APIScope {
	scopes := []APIScope{}

	for i, scope := range AllScopes {
		if mask&(1<<i) != 0 {
			scopes = append(scopes, scope)
		}
	}

	return scopes
}
