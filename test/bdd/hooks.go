package bdd

import (
	"context"

	"github.com/cucumber/godog"
)

// InitializeTestSuite runs once before all scenarios
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Suite-level setup if needed (none for pilot)
}

// InitializeScenario runs before each scenario and registers step definitions
func InitializeScenario(ctx *godog.ScenarioContext) {
	// Per-scenario setup: create fresh World
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		w := NewWorld()

		// For pilot, start HTTP server for all scenarios
		// TODO Phase B: Optimize to only start for @http/@ui tagged scenarios
		w.startHTTPServer()

		return withWorld(ctx, w), nil
	})

	// Per-scenario teardown
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		w := getWorld(ctx)
		w.Reset()
		return ctx, nil
	})

	// Register all step definitions
	RegisterCommonSteps(ctx)
	RegisterScheduleSteps(ctx)
	RegisterGroupSteps(ctx)
	RegisterHTTPSteps(ctx)
	RegisterAssignmentSteps(ctx) // Phase B
	RegisterAuthSteps(ctx)        // Phase C
	RegisterUserSteps(ctx)        // Phase C
}
