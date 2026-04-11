package bdd

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output:    colors.Colored(os.Stdout),
	Format:    "pretty",
	Paths:     []string{"features"},
	Strict:    true,
	Randomize: -1,
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = []string{"features"}

	// Run tests
	status := m.Run()

	os.Exit(status)
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t
	status := godog.TestSuite{
		Name:                 "deployment-tail-bdd",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &o,
	}.Run()
	if status != 0 {
		t.Fatalf("godog suite failed with status %d", status)
	}
}
