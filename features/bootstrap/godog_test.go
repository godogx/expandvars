package bootstrap

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/godogx/expandvars"
)

// Used by init().
//
//nolint:gochecknoglobals
var (
	runGoDogTests bool

	out = new(bytes.Buffer)
	opt = godog.Options{
		Strict: true,
		Output: out,
	}
)

// This has to run on init to define -godog flag, otherwise "undefined flag" error happens.
//
//nolint:gochecknoinits
func init() {
	flag.BoolVar(&runGoDogTests, "godog", false, "Set this flag is you want to run godog BDD tests")
	godog.BindCommandLineFlags("godog.", &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	pflag.Parse()

	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	if !runGoDogTests {
		t.Skip(`Missing "--godog" flag, skipping integration test.`)
	}

	stepExpander := expandvars.NewStepExpander(
		expandvars.Pairs{"NAME": "John"},
		expandvars.BeforeScenario(func() expandvars.Pairs {
			return expandvars.Pairs{
				"TIMESTAMP": fmt.Sprintf("%d", time.Now().UnixNano()),
			}
		}),
		expandvars.EnvExpander,
	)

	err := os.Setenv("FOOBAR", "foobar") //nolint: tenv
	assert.NoError(t, err)

	stepCtx := newStepCtx()

	RunSuite(t, "..", func(t *testing.T, ctx *godog.ScenarioContext) {
		t.Helper()

		stepExpander.RegisterContext(ctx)
		stepCtx.RegisterContext(t, ctx)
	})
}

func RunSuite(t *testing.T, path string, featureContext func(t *testing.T, ctx *godog.ScenarioContext)) {
	t.Helper()

	var paths []string

	files, err := ioutil.ReadDir(filepath.Clean(path))
	assert.NoError(t, err)

	paths = make([]string, 0, len(files))

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".feature") {
			paths = append(paths, filepath.Join(path, f.Name()))
		}
	}

	for _, path := range paths {
		path := path

		t.Run(path, func(t *testing.T) {
			opt.Paths = []string{path}
			suite := godog.TestSuite{
				Name:                 "Integration",
				TestSuiteInitializer: nil,
				ScenarioInitializer: func(s *godog.ScenarioContext) {
					featureContext(t, s)
				},
				Options: &opt,
			}
			status := suite.Run()

			if status != 0 {
				fmt.Println(out.String())
				assert.Failf(t, "one or more scenarios failed in feature: %s", path)
			}
		})
	}
}
