package api

import (
	"strconv"
	"strings"
)

// oidRole oid identifier used to store user roles
const oidRole string = "1.2.840.10070.8.1"

// permissions
var permissions = map[string][]string{
	"/WorkerService/Start":  {"full"},
	"/WorkerService/Stop":   {"full"},
	"/WorkerService/Query":  {"full", "read"},
	"/WorkerService/Stream": {"full", "read"},
}

func HasPermission(method string, roles []string) bool {
	permission, ok := permissions[method]
	if !ok {
		return false
	}
	for _, role := range roles {
		for _, value := range permission {
			if role == value {
				return true
			}
		}
	}
	return false
}

func IsOidRole(oid string) bool {
	return oidRole == oid
}

func ParseRoles(roles string) []string {
	return strings.Split(strings.TrimSpace(roles), ",")
}

func OidToString(oid []int) string {
	var strs []string
	for _, value := range oid {
		strs = append(strs, strconv.Itoa(value))
	}
	return strings.Join(strs, ".")
}
