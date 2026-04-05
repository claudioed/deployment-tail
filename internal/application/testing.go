package application

import (
	"context"

	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// MockRepository is a mock implementation of schedule.Repository for testing
type MockRepository struct {
	schedules map[string]*schedule.Schedule
	createErr error
	findErr   error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		schedules: make(map[string]*schedule.Schedule),
	}
}

func (m *MockRepository) Create(ctx context.Context, sch *schedule.Schedule) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.schedules[sch.ID().String()] = sch
	return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id schedule.ScheduleID) (*schedule.Schedule, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	sch, ok := m.schedules[id.String()]
	if !ok {
		return nil, schedule.ErrScheduleNotFound
	}
	return sch, nil
}

func (m *MockRepository) FindAll(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	var result []*schedule.Schedule
	for _, sch := range m.schedules {
		result = append(result, sch)
	}
	return result, nil
}

func (m *MockRepository) Update(ctx context.Context, sch *schedule.Schedule) error {
	if _, ok := m.schedules[sch.ID().String()]; !ok {
		return schedule.ErrScheduleNotFound
	}
	m.schedules[sch.ID().String()] = sch
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id schedule.ScheduleID, deletedBy user.UserID) error {
	if _, ok := m.schedules[id.String()]; !ok {
		return schedule.ErrScheduleNotFound
	}
	delete(m.schedules, id.String())
	return nil
}

func (m *MockRepository) FindUngrouped(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	return m.FindAll(ctx, filters)
}

// MockGroupRepository is a mock implementation of group.Repository for testing
type MockGroupRepository struct {
	groups           map[string]*group.Group
	scheduleGroups   map[string][]string // scheduleID -> []groupID
	createErr        error
	findErr          error
	duplicateNameErr bool
}

func NewMockGroupRepository() *MockGroupRepository {
	return &MockGroupRepository{
		groups:         make(map[string]*group.Group),
		scheduleGroups: make(map[string][]string),
	}
}

func (m *MockGroupRepository) Create(ctx context.Context, grp *group.Group) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.duplicateNameErr {
		return group.ErrDuplicateGroupName
	}
	m.groups[grp.ID().String()] = grp
	return nil
}

func (m *MockGroupRepository) FindAll(ctx context.Context, owner schedule.Owner) ([]*group.Group, error) {
	var result []*group.Group
	for _, grp := range m.groups {
		if grp.Owner().Equals(owner) {
			result = append(result, grp)
		}
	}
	return result, nil
}

func (m *MockGroupRepository) FindByID(ctx context.Context, id group.GroupID) (*group.Group, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	grp, ok := m.groups[id.String()]
	if !ok {
		return nil, group.ErrGroupNotFound
	}
	return grp, nil
}

func (m *MockGroupRepository) Update(ctx context.Context, grp *group.Group) error {
	if _, ok := m.groups[grp.ID().String()]; !ok {
		return group.ErrGroupNotFound
	}
	if m.duplicateNameErr {
		return group.ErrDuplicateGroupName
	}
	m.groups[grp.ID().String()] = grp
	return nil
}

func (m *MockGroupRepository) Delete(ctx context.Context, id group.GroupID) error {
	if _, ok := m.groups[id.String()]; !ok {
		return group.ErrGroupNotFound
	}
	delete(m.groups, id.String())
	delete(m.scheduleGroups, id.String())
	return nil
}

func (m *MockGroupRepository) AddSchedule(ctx context.Context, groupID group.GroupID, scheduleID schedule.ScheduleID, assignedBy string) error {
	if _, ok := m.groups[groupID.String()]; !ok {
		return group.ErrGroupNotFound
	}
	m.scheduleGroups[scheduleID.String()] = append(m.scheduleGroups[scheduleID.String()], groupID.String())
	return nil
}

func (m *MockGroupRepository) RemoveSchedule(ctx context.Context, groupID group.GroupID, scheduleID schedule.ScheduleID) error {
	if groups, ok := m.scheduleGroups[scheduleID.String()]; ok {
		var filtered []string
		for _, gid := range groups {
			if gid != groupID.String() {
				filtered = append(filtered, gid)
			}
		}
		m.scheduleGroups[scheduleID.String()] = filtered
	}
	return nil
}

func (m *MockGroupRepository) GetSchedulesInGroup(ctx context.Context, groupID group.GroupID) ([]schedule.ScheduleID, error) {
	var result []schedule.ScheduleID
	for schedID, groupIDs := range m.scheduleGroups {
		for _, gid := range groupIDs {
			if gid == groupID.String() {
				sid, _ := schedule.ParseScheduleID(schedID)
				result = append(result, sid)
				break
			}
		}
	}
	return result, nil
}

func (m *MockGroupRepository) GetGroupsForSchedule(ctx context.Context, scheduleID schedule.ScheduleID) ([]*group.Group, error) {
	var result []*group.Group
	if groupIDs, ok := m.scheduleGroups[scheduleID.String()]; ok {
		for _, gid := range groupIDs {
			if grp, ok := m.groups[gid]; ok {
				result = append(result, grp)
			}
		}
	}
	return result, nil
}

// MockUserRepository is a mock implementation of user.Repository for testing
type MockUserRepository struct {
	users              map[string]*user.User
	findByIDError      error
	findByGoogleIDErr  error
	createError        error
	updateError        error
	updateLastLoginErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*user.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	if m.createError != nil {
		return m.createError
	}
	m.users[u.ID().String()] = u
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	if m.findByIDError != nil {
		return nil, m.findByIDError
	}
	u, exists := m.users[id.String()]
	if !exists {
		return nil, user.ErrUserNotFound{ID: id.String(), SearchType: "id"}
	}
	return u, nil
}

func (m *MockUserRepository) FindByGoogleID(ctx context.Context, googleID user.GoogleID) (*user.User, error) {
	if m.findByGoogleIDErr != nil {
		return nil, m.findByGoogleIDErr
	}
	for _, u := range m.users {
		if u.GoogleID().Equals(googleID) {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound{GoogleID: googleID.String(), SearchType: "google_id"}
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	for _, u := range m.users {
		if u.Email().Equals(email) {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound{Email: email.String(), SearchType: "email"}
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.users[u.ID().String()] = u
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, filters user.ListFilters) ([]*user.User, error) {
	var result []*user.User
	for _, u := range m.users {
		if filters.Role != nil && !u.Role().Equals(*filters.Role) {
			continue
		}
		result = append(result, u)
	}
	return result, nil
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID user.UserID, role user.Role) error {
	u, exists := m.users[userID.String()]
	if !exists {
		return user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
	}
	u.UpdateRole(role)
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID user.UserID) error {
	if m.updateLastLoginErr != nil {
		return m.updateLastLoginErr
	}
	u, exists := m.users[userID.String()]
	if !exists {
		return user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
	}
	u.RecordLogin()
	return nil
}
