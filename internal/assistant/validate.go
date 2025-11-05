package assistant

import (
	"errors"
	"strings"
)

var (
	ErrEmptyID      = errors.New("assistant ID cannot be empty")
	ErrInvalidChars = errors.New("assistant ID contains invalid characters")
	ErrReservedName = errors.New("assistant ID cannot be '.' or '..'")
	ErrTooLong      = errors.New("assistant ID exceeds 255 characters")
)

// invalidChars contains characters not allowed in directory names.
var invalidChars = []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}

// ValidateID checks if the given ID is a valid directory name.
func ValidateID(id string) error {
	if id == "" {
		return ErrEmptyID
	}
	if id == "." || id == ".." {
		return ErrReservedName
	}
	if len(id) > 255 {
		return ErrTooLong
	}
	for _, char := range invalidChars {
		if strings.ContainsRune(id, char) {
			return ErrInvalidChars
		}
	}
	return nil
}
