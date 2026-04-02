package schedule

import "testing"

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    Status
		wantErr bool
	}{
		{
			name:    "parse created",
			value:   "created",
			want:    StatusCreated,
			wantErr: false,
		},
		{
			name:    "parse approved",
			value:   "approved",
			want:    StatusApproved,
			wantErr: false,
		},
		{
			name:    "parse denied",
			value:   "denied",
			want:    StatusDenied,
			wantErr: false,
		},
		{
			name:    "invalid status",
			value:   "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStatus(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{
			name:   "created to string",
			status: StatusCreated,
			want:   "created",
		},
		{
			name:   "approved to string",
			status: StatusApproved,
			want:   "approved",
		},
		{
			name:   "denied to string",
			status: StatusDenied,
			want:   "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		from    Status
		to      Status
		wantErr bool
	}{
		{
			name:    "created to approved - valid",
			from:    StatusCreated,
			to:      StatusApproved,
			wantErr: false,
		},
		{
			name:    "created to denied - valid",
			from:    StatusCreated,
			to:      StatusDenied,
			wantErr: false,
		},
		{
			name:    "approved to denied - invalid",
			from:    StatusApproved,
			to:      StatusDenied,
			wantErr: true,
		},
		{
			name:    "denied to approved - invalid",
			from:    StatusDenied,
			to:      StatusApproved,
			wantErr: true,
		},
		{
			name:    "approved to created - invalid",
			from:    StatusApproved,
			to:      StatusCreated,
			wantErr: true,
		},
		{
			name:    "denied to created - invalid",
			from:    StatusDenied,
			to:      StatusCreated,
			wantErr: true,
		},
		{
			name:    "created to created - same status",
			from:    StatusCreated,
			to:      StatusCreated,
			wantErr: true,
		},
		{
			name:    "approved to approved - same status",
			from:    StatusApproved,
			to:      StatusApproved,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.from.CanTransitionTo(tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Status.CanTransitionTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
