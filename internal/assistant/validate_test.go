package assistant

import "testing"

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"valid simple", "my-assistant", nil},
		{"valid with spaces", "My Assistant", nil},
		{"valid with numbers", "assistant123", nil},
		{"empty", "", ErrEmptyID},
		{"dot", ".", ErrReservedName},
		{"double dot", "..", ErrReservedName},
		{"contains slash", "foo/bar", ErrInvalidChars},
		{"contains backslash", "foo\\bar", ErrInvalidChars},
		{"contains colon", "foo:bar", ErrInvalidChars},
		{"contains asterisk", "foo*bar", ErrInvalidChars},
		{"contains question mark", "foo?bar", ErrInvalidChars},
		{"contains quote", "foo\"bar", ErrInvalidChars},
		{"contains less than", "foo<bar", ErrInvalidChars},
		{"contains greater than", "foo>bar", ErrInvalidChars},
		{"contains pipe", "foo|bar", ErrInvalidChars},
		{"too long", string(make([]byte, 256)), ErrTooLong},
		{"max length", string(make([]byte, 255)), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if err != tt.wantErr {
				t.Errorf("ValidateID(%q) = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}
