package store

type LimitAction string

const (
	LimitLimitAction   LimitAction = "limit"
	WarningLimitAction LimitAction = "warning"
)

type LimitScope string

const (
	AccountLimitScope     LimitScope = "account"
	EnvironmentLimitScope LimitScope = "environment"
	ShareLimitScope       LimitScope = "share"
)

type PermissionMode string

const (
	OpenPermissionMode   PermissionMode = "open"
	ClosedPermissionMode PermissionMode = "closed"
)
