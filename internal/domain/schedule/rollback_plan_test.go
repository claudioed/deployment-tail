package schedule

import (
	"strings"
	"testing"
)

func TestNewRollbackPlan(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid rollback plan",
			value:   "Revert to v1.2.3 using git reset",
			wantErr: false,
		},
		{
			name:    "empty rollback plan",
			value:   "",
			wantErr: false,
		},
		{
			name:    "multiline rollback plan",
			value:   "Step 1: Stop service\nStep 2: Restore backup\nStep 3: Start service",
			wantErr: false,
		},
		{
			name:    "rollback plan with special characters",
			value:   "kubectl rollback deployment/app --to-revision=2",
			wantErr: false,
		},
		{
			name:    "rollback plan too long",
			value:   strings.Repeat("a", 5001),
			wantErr: true,
		},
		{
			name:    "rollback plan at max length",
			value:   strings.Repeat("a", 5000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := NewRollbackPlan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRollbackPlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && plan.String() != tt.value {
				t.Errorf("NewRollbackPlan() = %v, want %v", plan.String(), tt.value)
			}
		})
	}
}

func TestRollbackPlan_IsEmpty(t *testing.T) {
	emptyPlan, _ := NewRollbackPlan("")
	nonEmptyPlan, _ := NewRollbackPlan("some plan")

	if !emptyPlan.IsEmpty() {
		t.Error("Expected empty plan to be empty")
	}

	if nonEmptyPlan.IsEmpty() {
		t.Error("Expected non-empty plan to not be empty")
	}
}

func TestRollbackPlan_Equals(t *testing.T) {
	plan1, _ := NewRollbackPlan("plan A")
	plan2, _ := NewRollbackPlan("plan A")
	plan3, _ := NewRollbackPlan("plan B")

	if !plan1.Equals(plan2) {
		t.Error("Expected plan1 to equal plan2")
	}

	if plan1.Equals(plan3) {
		t.Error("Expected plan1 to not equal plan3")
	}
}
