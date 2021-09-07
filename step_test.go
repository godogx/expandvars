package expandvars_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	"github.com/godogx/expandvars"
)

func TestExpandStep(t *testing.T) {
	t.Parallel()

	expanders := []interface{}{
		// Raw replacer.
		strings.NewReplacer("$TO", "Berlin"),
		// Our expanders.
		expandvars.Pairs{
			"HUSBAND": "John",
		},
		func() expandvars.Pairs {
			return expandvars.Pairs{
				"WIFE": "Jane",
			}
		},
		func() expandvars.Expander {
			return func(s string) string {
				return strings.ReplaceAll(s, "$DURATION", "and stay there for 3 days")
			}
		},
		func(s string) string {
			return strings.ReplaceAll(s, "$FROM", "Paris")
		},
		expandvars.Expander(func(s string) string {
			return strings.ReplaceAll(s, "$TRANSPORT", "by bus")
		}),
		// Os.
		expandvars.EnvExpander,
	}

	// Set os env.
	assert.NoError(t, os.Setenv("GREETINGS", "Hi Dave"))

	defer func() {
		_ = os.Unsetenv("GREETINGS") // nolint:errcheck
	}()

	step := &godog.Step{Text: "$GREETINGS, $HUSBAND & $WIFE are going from $FROM to $TO $TRANSPORT $DURATION"}
	expected := "Hi Dave, John & Jane are going from Paris to Berlin by bus and stay there for 3 days"

	expandvars.ExpandStep(step, expanders...)

	assert.Equal(t, expected, step.Text)
}
