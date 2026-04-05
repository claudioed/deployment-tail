package group

import (
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

// Group is the aggregate root for custom schedule groups
type Group struct {
	id          GroupID
	name        GroupName
	description Description
	owner       schedule.Owner
	createdAt   time.Time
	updatedAt   time.Time
}

// NewGroup creates a new group
func NewGroup(
	name GroupName,
	description Description,
	owner schedule.Owner,
) (*Group, error) {
	now := time.Now().UTC()

	return &Group{
		id:          NewGroupID(),
		name:        name,
		description: description,
		owner:       owner,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstitute recreates a group from storage
func Reconstitute(
	id GroupID,
	name GroupName,
	description Description,
	owner schedule.Owner,
	createdAt time.Time,
	updatedAt time.Time,
) *Group {
	return &Group{
		id:          id,
		name:        name,
		description: description,
		owner:       owner,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Rename updates the group name
func (g *Group) Rename(name GroupName) error {
	g.name = name
	g.updatedAt = time.Now().UTC()
	return nil
}

// UpdateDescription updates the group description
func (g *Group) UpdateDescription(description Description) error {
	g.description = description
	g.updatedAt = time.Now().UTC()
	return nil
}

// Getters

func (g *Group) ID() GroupID {
	return g.id
}

func (g *Group) Name() GroupName {
	return g.name
}

func (g *Group) Description() Description {
	return g.description
}

func (g *Group) Owner() schedule.Owner {
	return g.owner
}

func (g *Group) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Group) UpdatedAt() time.Time {
	return g.updatedAt
}
