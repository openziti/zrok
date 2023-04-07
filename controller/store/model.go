package store

type LimitJournalAction string

const (
	LimitAction   LimitJournalAction = "limit"
	WarningAction LimitJournalAction = "warning"
	ClearAction   LimitJournalAction = "clear"
)
