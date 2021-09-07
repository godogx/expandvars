# Variables Expander for Cucumber Steps

[![GitHub Releases](https://img.shields.io/github/v/release/godogx/expandvars)](https://github.com/godogx/expandvars/releases/latest)
[![Build Status](https://github.com/godogx/expandvars/actions/workflows/test.yaml/badge.svg)](https://github.com/godogx/expandvars/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/godogx/expandvars/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/godogx/expandvars)
[![Go Report Card](https://goreportcard.com/badge/github.com/godogx/expandvars)](https://goreportcard.com/report/github.com/godogx/expandvars)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/godogx/expandvars)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

A lifesaver expander for [`cucumber/godog`](https://github.com/cucumber/godog) because, sometimes, you have to use variables in your steps.

## Prerequisites

- `Go >= 1.16`

## Install

```bash
go get github.com/godogx/expandvars
```

## Usage

Initiate a new `StepExpander` with `expandvars.NewStepExpander()` then add it to `ScenarioInitializer` by
calling `StepExpander.RegisterContext(*testing.T, *godog.ScenarioContext)`

```go
package main

import (
    "fmt"
    "math/rand"
    "strings"
    "testing"

    "github.com/cucumber/godog"
    "github.com/godogx/expandvars"
    "github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
    expander := expandvars.NewStepExpander(
        strings.NewReplacer("$TO", "Berlin"),
        expandvars.Pairs{
            "HUSBAND": "John",
        },
        func() expandvars.Pairs {
            return expandvars.Pairs{
                "RAND": fmt.Sprintf("%d", rand.Int63()),
            }
        },
        expandvars.BeforeScenario(func() expandvars.Pairs {
            return expandvars.Pairs{
                "SCENARIO_RAND": fmt.Sprintf("%d", rand.Int63()),
            }
        }),
        func(s string) string {
            return strings.ReplaceAll(s, "$FROM", "Paris")
        },
        expandvars.Expander(func(s string) string {
            return strings.ReplaceAll(s, "$TRANSPORT", "by bus")
        }),
        // OS env vars.
        expandvars.EnvExpander,
    )

    suite := godog.TestSuite{
        Name: "Integration",
        ScenarioInitializer: func(ctx *godog.ScenarioContext) {
            expander.RegisterContext(ctx)
        },
        Options: &godog.Options{
            Strict:    true,
            Randomize: rand.Int63(),
        },
    }

    // Run the suite.
}
```

In your tests, just use `$VARIABLE_NAME` in the step or the argument, like this:

```gherkin
    Scenario: var is replaced
        Given var NAME is replaced in step text: $NAME

        Then step text is:
        """
        map var NAME is replaced in step text: John
        """

        Given var NAME is replaced in step argument (string)
        """
        NAME=$NAME
        """

        Then step argument is a string:
        """
        NAME=John
        """

        Given env var NAME is replaced in step argument (table)
            | col 1   | col 2 | col 3   |
            | value 1 | $NAME | value 3 |

        Then step argument is a table:
            | col 1   | col 2 | col 3   |
            | value 1 | John  | value 3 |
```

```gherkin
    Scenario: .github files
        Then there should be only these files in "$TEST_DIR/.github":
        """
        - workflows:
            - golangci-lint.yaml
            - test.yaml
        """
```

### Expanders

The expanders could be any of these:

1. A `Replacer` interface

```go
type Replacer interface {
    Replace(string) string
}
```

2. A `Replacer` `func(string) string` function. <br/><br/>
   For example, you could use `os.ExpandEnv` or its alias `expandvars.EnvExpander`

3. A map or vars (without the `$`) `map[string]string`

```go
var _ = expandvars.NewStepExpander(expandvars.Pairs{
    "HUSBAND": "John",
    "WIFE":    "Jane",
})
```

4. A provider that provides a map of vars (without the `$`) `map[string]string`. The provider will be called every step.

```go
var _ = expandvars.NewStepExpander(func() expandvars.Pairs {
    return map[string]string{
        "RAND": fmt.Sprintf("%d", rand.Int63()),
    }
})
```

5. A `BeforeScenario` provides a map of vars (without the `$`) `map[string]string`. The provider will be called only once before every scenario.

**Note**: If you need `expandvars.EnvExpander` or `os.ExpandEnv`, put it in the end of the chain. Because it replaces not-found vars with empty strings, other
expanders won't have a chance to do their jobs if you put it in the beginning.
