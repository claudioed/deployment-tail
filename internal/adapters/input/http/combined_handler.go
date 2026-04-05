package http

import (
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/infrastructure/oauth"
)

// CombinedHandler combines all API handlers
type CombinedHandler struct {
	*ScheduleHandler
	*GroupHandler
	*UserHandler
	*AuthHandler
}

// NewCombinedHandler creates a new combined handler
func NewCombinedHandler(
	scheduleService input.ScheduleService,
	groupService input.GroupService,
	userService input.UserService,
	googleClient *oauth.GoogleClient,
) *CombinedHandler {
	return &CombinedHandler{
		ScheduleHandler: NewScheduleHandler(scheduleService, groupService, userService),
		GroupHandler:    NewGroupHandler(groupService, scheduleService),
		UserHandler:     NewUserHandler(userService),
		AuthHandler:     NewAuthHandler(userService, googleClient),
	}
}
