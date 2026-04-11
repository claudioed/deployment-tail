package http

import (
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/infrastructure/oauth"
)

// CombinedHandler combines all API handlers
type CombinedHandler struct {
	*ScheduleHandler
	*GroupHandler
	*UserHandler
	*AuthHandler
	*ServiceHandler
}

// NewCombinedHandler creates a new combined handler
func NewCombinedHandler(
	scheduleService input.ScheduleService,
	groupService input.GroupService,
	userService input.UserService,
	serviceService *application.ServiceService,
	googleClient *oauth.GoogleClient,
) *CombinedHandler {
	return &CombinedHandler{
		ScheduleHandler: NewScheduleHandler(scheduleService, groupService, userService),
		GroupHandler:    NewGroupHandler(groupService, scheduleService),
		UserHandler:     NewUserHandler(userService),
		AuthHandler:     NewAuthHandler(userService, googleClient),
		ServiceHandler:  NewServiceHandler(serviceService),
	}
}
