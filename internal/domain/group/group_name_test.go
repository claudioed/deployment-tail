package group

import (
	"strings"
	"testing"
)

func TestNewGroupName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			value:   "Project Alpha",
			wantErr: false,
		},
		{
			name:    "valid name with special characters",
			value:   "Team-Backend_2024",
			wantErr: false,
		},
		{
			name:    "valid name at max length",
			value:   strings.Repeat("a", 100),
			wantErr: false,
		},
		{
			name:    "empty name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "name too long",
			value:   strings.Repeat("a", 101),
			wantErr: true,
		},
		{
			name:    "name with unicode",
			value:   "Projet Français",
			wantErr: false,
		},
		{
			name:    "single character",
			value:   "A",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewGroupName(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGroupName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && name.String() != tt.value {
				t.Errorf("NewGroupName() = %v, want %v", name.String(), tt.value)
			}
		})
	}
}

func TestGroupName_Equals(t *testing.T) {
	name1, _ := NewGroupName("Project Alpha")
	name2, _ := NewGroupName("Project Alpha")
	name3, _ := NewGroupName("Project Beta")

	if !name1.Equals(name2) {
		t.Error("Expected name1 to equal name2")
	}

	if name1.Equals(name3) {
		t.Error("Expected name1 to not equal name3")
	}
}
