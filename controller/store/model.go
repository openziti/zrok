package store

type LimitAction string

const (
	LimitLimitAction   LimitAction = "limit"
	WarningLimitAction LimitAction = "warning"
)

type PermissionMode string

const (
	OpenPermissionMode   PermissionMode = "open"
	ClosedPermissionMode PermissionMode = "closed"
)
