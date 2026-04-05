package group

import (
	"strings"
	"testing"
)

func TestNewDescription(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "empty description",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid short description",
			value:   "All schedules for Project Alpha",
			wantErr: false,
		},
		{
			name:    "valid description at max length",
			value:   strings.Repeat("a", 500),
			wantErr: false,
		},
		{
			name:    "description too long",
			value:   strings.Repeat("a", 501),
			wantErr: true,
		},
		{
			name:    "description with newlines",
			value:   "First line\nSecond line",
			wantErr: false,
		},
		{
			name:    "description with unicode",
			value:   "Description avec des caractères français",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, err := NewDescription(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && desc.String() != tt.value {
				t.Errorf("NewDescription() = %v, want %v", desc.String(), tt.value)
			}
		})
	}
}

func TestDescription_IsEmpty(t *testing.T) {
	emptyDesc, _ := NewDescription("")
	if !emptyDesc.IsEmpty() {
		t.Error("Expected empty description to report IsEmpty() = true")
	}

	nonEmptyDesc, _ := NewDescription("Some text")
	if nonEmptyDesc.IsEmpty() {
		t.Error("Expected non-empty description to report IsEmpty() = false")
	}
}

func TestDescription_Value(t *testing.T) {
	text := "Test description"
	desc, _ := NewDescription(text)

	if desc.Value() != text {
		t.Errorf("Value() = %v, want %v", desc.Value(), text)
	}
}
