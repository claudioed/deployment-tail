package schedule

import (
	"strings"
	"testing"
)

func TestNewOwner(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid alphanumeric owner",
			value:   "user123",
			wantErr: false,
		},
		{
			name:    "valid email owner",
			value:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid with dots and hyphens",
			value:   "john.doe-dev",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			value:   "user_name",
			wantErr: false,
		},
		{
			name:    "empty owner",
			value:   "",
			wantErr: true,
		},
		{
			name:    "owner too long",
			value:   strings.Repeat("a", 256),
			wantErr: true,
		},
		{
			name:    "invalid characters (spaces)",
			value:   "user name",
			wantErr: true,
		},
		{
			name:    "invalid characters (special chars)",
			value:   "user!name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, err := NewOwner(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOwner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && owner.String() != tt.value {
				t.Errorf("NewOwner() = %v, want %v", owner.String(), tt.value)
			}
		})
	}
}

func TestOwner_Equals(t *testing.T) {
	owner1, _ := NewOwner("user1")
	owner2, _ := NewOwner("user1")
	owner3, _ := NewOwner("user2")

	if !owner1.Equals(owner2) {
		t.Error("Expected owner1 to equal owner2")
	}

	if owner1.Equals(owner3) {
		t.Error("Expected owner1 to not equal owner3")
	}
}
