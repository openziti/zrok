package store

type LimitJournalAction string

const (
	LimitAction   LimitJournalAction = "limit"
	WarningAction LimitJournalAction = "warning"
	ClearAction   LimitJournalAction = "clear"
)

type PermissionMode string

const (
	OpenPermissionMode   PermissionMode = "open"
	ClosedPermissionMode PermissionMode = "closed"
)
