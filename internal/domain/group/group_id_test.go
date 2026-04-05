package group

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewGroupID(t *testing.T) {
	id := NewGroupID()

	// Should be a valid UUID
	if _, err := uuid.Parse(id.String()); err != nil {
		t.Errorf("NewGroupID() generated invalid UUID: %v", err)
	}

	// Should generate unique IDs
	id2 := NewGroupID()
	if id.Equals(id2) {
		t.Error("Expected NewGroupID() to generate unique IDs")
	}
}

func TestParseGroupID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			value:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "invalid UUID format",
			value:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "partial UUID",
			value:   "550e8400",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseGroupID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGroupID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id.String() != tt.value {
				t.Errorf("ParseGroupID() = %v, want %v", id.String(), tt.value)
			}
		})
	}
}

func TestGroupID_Equals(t *testing.T) {
	id1, _ := ParseGroupID("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := ParseGroupID("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := ParseGroupID("660e8400-e29b-41d4-a716-446655440000")

	if !id1.Equals(id2) {
		t.Error("Expected id1 to equal id2")
	}

	if id1.Equals(id3) {
		t.Error("Expected id1 to not equal id3")
	}
}
