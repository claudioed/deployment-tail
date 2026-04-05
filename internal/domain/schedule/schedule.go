package schedule

import (
	"fmt"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Schedule is the aggregate root for deployment schedules
type Schedule struct {
	id           ScheduleID
	scheduledAt  ScheduledTime
	service      ServiceName
	environments []Environment
	description  Description
	owners       []Owner
	status       Status
	rollbackPlan RollbackPlan
	createdBy    user.UserID // User who created the schedule
	updatedBy    user.UserID // User who last updated the schedule
	createdAt    time.Time
	updatedAt    time.Time
}

// NewSchedule creates a new schedule
func NewSchedule(
	scheduledAt ScheduledTime,
	service ServiceName,
	environments []Environment,
	description Description,
	owners []Owner,
	rollbackPlan RollbackPlan,
	createdBy user.UserID,
) (*Schedule, error) {
	// Validate at least one owner
	if len(owners) == 0 {
		return nil, fmt.Errorf("at least one owner is required")
	}

	// Validate at least one environment
	if len(environments) == 0 {
		return nil, fmt.Errorf("at least one environment is required")
	}

	// Deduplicate owners
	uniqueOwners := deduplicateOwners(owners)

	// Deduplicate environments
	uniqueEnvironments := deduplicateEnvironments(environments)

	now := time.Now().UTC()

	return &Schedule{
		id:           NewScheduleID(),
		scheduledAt:  scheduledAt,
		service:      service,
		environments: uniqueEnvironments,
		description:  description,
		owners:       uniqueOwners,
		status:       StatusCreated, // New schedules start with created status
		rollbackPlan: rollbackPlan,
		createdBy:    createdBy,
		updatedBy:    createdBy, // Initially same as createdBy
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Reconstitute recreates a schedule from storage
func Reconstitute(
	id ScheduleID,
	scheduledAt ScheduledTime,
	service ServiceName,
	environments []Environment,
	description Description,
	owners []Owner,
	status Status,
	rollbackPlan RollbackPlan,
	createdBy user.UserID,
	updatedBy user.UserID,
	createdAt time.Time,
	updatedAt time.Time,
) *Schedule {
	return &Schedule{
		id:           id,
		scheduledAt:  scheduledAt,
		service:      service,
		environments: environments,
		description:  description,
		owners:       owners,
		status:       status,
		rollbackPlan: rollbackPlan,
		createdBy:    createdBy,
		updatedBy:    updatedBy,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Update updates the schedule fields
func (s *Schedule) Update(
	scheduledAt *ScheduledTime,
	service *ServiceName,
	environments *[]Environment,
	description *Description,
	owners *[]Owner,
	rollbackPlan *RollbackPlan,
	updatedBy user.UserID,
) error {
	if scheduledAt != nil {
		s.scheduledAt = *scheduledAt
	}
	if service != nil {
		s.service = *service
	}
	if environments != nil {
		if len(*environments) == 0 {
			return fmt.Errorf("at least one environment is required")
		}
		// Validate each environment
		for _, env := range *environments {
			if !env.IsValid() {
				return fmt.Errorf("invalid environment: %s", env)
			}
		}
		s.environments = deduplicateEnvironments(*environments)
	}
	if owners != nil {
		if len(*owners) == 0 {
			return fmt.Errorf("at least one owner is required")
		}
		s.owners = deduplicateOwners(*owners)
	}
	if description != nil {
		s.description = *description
	}
	if rollbackPlan != nil {
		s.rollbackPlan = *rollbackPlan
	}
	s.updatedBy = updatedBy
	s.updatedAt = time.Now().UTC()
	return nil
}

// Approve changes the schedule status to approved
func (s *Schedule) Approve() error {
	if err := s.status.CanTransitionTo(StatusApproved); err != nil {
		return err
	}
	s.status = StatusApproved
	s.updatedAt = time.Now().UTC()
	return nil
}

// Deny changes the schedule status to denied
func (s *Schedule) Deny() error {
	if err := s.status.CanTransitionTo(StatusDenied); err != nil {
		return err
	}
	s.status = StatusDenied
	s.updatedAt = time.Now().UTC()
	return nil
}

// Getters

func (s *Schedule) ID() ScheduleID {
	return s.id
}

func (s *Schedule) ScheduledAt() ScheduledTime {
	return s.scheduledAt
}

func (s *Schedule) Service() ServiceName {
	return s.service
}

func (s *Schedule) Environments() []Environment {
	return s.environments
}

func (s *Schedule) Description() Description {
	return s.description
}

func (s *Schedule) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Schedule) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Schedule) Owners() []Owner {
	return s.owners
}

func (s *Schedule) Status() Status {
	return s.status
}

func (s *Schedule) RollbackPlan() RollbackPlan {
	return s.rollbackPlan
}

func (s *Schedule) CreatedBy() user.UserID {
	return s.createdBy
}

func (s *Schedule) UpdatedBy() user.UserID {
	return s.updatedBy
}

// AddOwner adds an owner to the schedule if not already present
func (s *Schedule) AddOwner(owner Owner) {
	// Check if owner already exists
	for _, existing := range s.owners {
		if existing.Equals(owner) {
			return // Already exists, don't add duplicate
		}
	}
	s.owners = append(s.owners, owner)
	s.updatedAt = time.Now().UTC()
}

// RemoveOwner removes an owner from the schedule
func (s *Schedule) RemoveOwner(owner Owner) error {
	// Can't remove last owner
	if len(s.owners) == 1 {
		return fmt.Errorf("cannot remove last owner")
	}

	for i, existing := range s.owners {
		if existing.Equals(owner) {
			s.owners = append(s.owners[:i], s.owners[i+1:]...)
			s.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("owner not found")
}

// AddEnvironment adds an environment to the schedule if not already present
func (s *Schedule) AddEnvironment(env Environment) {
	// Check if environment already exists
	for _, existing := range s.environments {
		if existing == env {
			return // Already exists, don't add duplicate
		}
	}
	s.environments = append(s.environments, env)
	s.updatedAt = time.Now().UTC()
}

// RemoveEnvironment removes an environment from the schedule
func (s *Schedule) RemoveEnvironment(env Environment) error {
	// Can't remove last environment
	if len(s.environments) == 1 {
		return fmt.Errorf("cannot remove last environment")
	}

	for i, existing := range s.environments {
		if existing == env {
			s.environments = append(s.environments[:i], s.environments[i+1:]...)
			s.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("environment not found")
}

// deduplicateOwners removes duplicate owners from a slice
func deduplicateOwners(owners []Owner) []Owner {
	seen := make(map[string]bool)
	result := []Owner{}
	for _, owner := range owners {
		key := owner.String()
		if !seen[key] {
			seen[key] = true
			result = append(result, owner)
		}
	}
	return result
}

// deduplicateEnvironments removes duplicate environments from a slice
func deduplicateEnvironments(environments []Environment) []Environment {
	seen := make(map[Environment]bool)
	result := []Environment{}
	for _, env := range environments {
		if !seen[env] {
			seen[env] = true
			result = append(result, env)
		}
	}
	return result
}
