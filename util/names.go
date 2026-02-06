package util

import (
	"regexp"

	goaway "github.com/TwiN/go-away"
)

// IsValidShareToken ensures that the string represents a valid unique name. Lowercase alphanumeric only. 3-32 characters.
func IsValidShareToken(uniqueName string) bool {
	match, err := regexp.Match("^[a-z0-9-]{3,32}$", []byte(uniqueName))
	if err != nil {
		return false
	}
	if match && goaway.IsProfane(uniqueName) {
		return false
	}
	return match
}

// IsValidNameInNamespace ensures that the string represents a valid name in a namespace. Lowercase alphanumeric only. 3-64 characters.
func IsValidNameInNamespace(uniqueName string) bool {
	match, err := regexp.Match("^[a-z0-9-]{3,63}$", []byte(uniqueName))
	if err != nil {
		return false
	}
	if match && goaway.IsProfane(uniqueName) {
		return false
	}
	return match
}
