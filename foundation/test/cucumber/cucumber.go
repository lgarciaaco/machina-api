// Package cucumber allows you to use cucumber to execute Gherkin based
// BDD test scenarios with some helpful API testing step implementations.
//
// Some steps allow you store variables or use those variables.  The variables
// are scoped to the Scenario.  The http response state is stored in the users
// session.  Switching users will switch the session.  Scenarios are executed
// concurrently.  The same user can be logged into two scenarios, but each scenario
// has a different session.
//
// Note: be careful using the same user/organization across different scenarios since
// they will likely see unexpected API mutations done in the other scenarios.
//
// Using in a test
//  func TestMain(m *testing.M) {
//
//	ocmServer := mocks.NewMockConfigurableServerBuilder().Build()
//	defer ocmServer.close()
//
//	h, _, teardown := test.RegisterIntegration(&testing.T{}, ocmServer)
//	defer teardown()
//
//	cucumber.TestMain(h)
//
//}
package cucumber

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"

	"github.com/lgarciaaco/machina-api/business/sys/auth"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/itchyny/gojq"
	"github.com/pmezard/go-difflib/difflib"
)

// TestSuite holds the sate global to all the test scenarios.
// It is accessed concurrently from all test scenarios.
type TestSuite struct {
	APIURL        string
	Mu            sync.Mutex
	DB            *sqlx.DB
	Logger        *zap.SugaredLogger
	authenticator *auth.Auth
	users         map[string]*TestUser
}

// TestUser represents a user that can login to the system.  The same users are shared by
// the different test scenarios.
type TestUser struct {
	Name  string
	Token string
	Ctx   context.Context
	Mu    sync.Mutex
}

// TestScenario holds that state of single scenario.  It is not accessed
// concurrently.
type TestScenario struct {
	Suite           *TestSuite
	DB              *sqlx.DB
	CurrentUser     string
	PathPrefix      string
	sessions        map[string]*TestSession
	Variables       map[string]interface{}
	hasTestCaseLock bool
}

func (s *TestScenario) User() *TestUser {
	s.Suite.Mu.Lock()
	defer s.Suite.Mu.Unlock()
	return s.Suite.users[s.CurrentUser]
}

func (s *TestScenario) Session() *TestSession {
	result := s.sessions[s.CurrentUser]
	if result == nil {
		result = &TestSession{
			TestUser: s.User(),
			Client:   &http.Client{},
			Header:   http.Header{},
		}
		s.sessions[s.CurrentUser] = result
	}
	return result
}

func (s *TestScenario) JSONMustMatch(actual, expected string, expand bool) error {

	var actualParsed interface{}
	err := json.Unmarshal([]byte(actual), &actualParsed)
	if err != nil {
		return fmt.Errorf("error parsing actual json: %v\njson was:\n%s", err, actual)
	}

	var expectedParsed interface{}
	expanded := expected
	if expand {
		expanded, err = s.Expand(expected)
		if err != nil {
			return err
		}
	}
	if err := json.Unmarshal([]byte(expanded), &expectedParsed); err != nil {
		return fmt.Errorf("error parsing expected json: %v\njson was:\n%s", err, expanded)
	}

	if !reflect.DeepEqual(expectedParsed, actualParsed) {
		expected, _ := json.MarshalIndent(expectedParsed, "", "  ")
		actual, _ := json.MarshalIndent(actualParsed, "", "  ")

		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(expected)),
			B:        difflib.SplitLines(string(actual)),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})
		return fmt.Errorf("actual does not match expected, diff:\n%s", diff)
	}

	return nil
}

// Expand replaces ${var} or $var in the string based on saved Variables in the session/test scenario.
func (s *TestScenario) Expand(value string) (result string, rerr error) {
	session := s.Session()
	return os.Expand(value, func(name string) string {

		arrayResponse := strings.HasPrefix(name, "response[")
		if strings.HasPrefix(name, "response.") || arrayResponse {

			selector := "." + name
			query, err := gojq.Parse(selector)
			if err != nil {
				rerr = err
				return ""
			}

			j, err := session.RespJSON()
			if err != nil {
				rerr = err
				return ""
			}

			j = map[string]interface{}{
				"response": j,
			}

			iter := query.Run(j)
			if next, found := iter.Next(); found {
				switch next := next.(type) {
				case string:
					return next
				case int:
					return fmt.Sprintf("%d", next)
				case float64:
					return fmt.Sprintf("%f", next)
				case float32:
					return fmt.Sprintf("%f", next)
				case nil:
					rerr = fmt.Errorf("field ${%s} not found in json response:\n%s", name, string(session.RespBytes))
					return ""
				case error:
					rerr = fmt.Errorf("failed to evaluate selection: %s: %v", name, next)
					return ""
				default:
					return fmt.Sprintf("%s", next)
				}
			} else {
				rerr = fmt.Errorf("field ${%s} not found in json response:\n%s", name, string(session.RespBytes))
				return ""
			}
		}
		value, found := s.Variables[name]
		if !found {
			return ""
		}
		return fmt.Sprint(value)
	}), rerr
}

// TestSession holds the http context for a user kinda like a browser.  Each scenario
// had a different session even if using the same user.
type TestSession struct {
	TestUser          *TestUser
	Client            *http.Client
	Resp              *http.Response
	Ctx               context.Context
	RespBytes         []byte
	respJSON          interface{}
	Header            http.Header
	EventStream       bool
	EventStreamEvents chan interface{}
	Debug             bool
}

// RespJSON returns the last http response body as json
func (s *TestSession) RespJSON() (interface{}, error) {
	if s.respJSON == nil {
		if err := json.Unmarshal(s.RespBytes, &s.respJSON); err != nil {
			return nil, fmt.Errorf("error parsing json response: %v\nbody: %s", err, string(s.RespBytes))
		}

		if s.Debug {
			fmt.Println("response json:")
			e := json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			_ = e.Encode(s.respJSON)
			fmt.Println("")
		}
	}
	return s.respJSON, nil
}

func (s *TestSession) SetRespBytes(bytes []byte) {
	s.RespBytes = bytes
	s.respJSON = nil
}

// StepModules is the list of functions used to add steps to a godog.ScenarioContext, you can
// add more to this list if you need test TestSuite specific steps.
var StepModules []func(ctx *godog.ScenarioContext, s *TestScenario)

func (s *TestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ts := &TestScenario{
		Suite:     s,
		sessions:  map[string]*TestSession{},
		Variables: map[string]interface{}{},
		DB:        s.DB,
	}

	for _, module := range StepModules {
		module(ctx, ts)
	}
}

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "progress", // can define default values
	Paths:       []string{"features"},
	Randomize:   time.Now().UTC().UnixNano(), // randomize TestScenario execution order
	Concurrency: 10,
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

// RunSuite runs the scenarios found in the "$path" directory.  If t is not nil, it
// also runs it's tests.
func RunSuite(path string, logger *zap.SugaredLogger, db *sqlx.DB, authenticator *auth.Auth, t *testing.T) {
	s := &TestSuite{
		APIURL:        "http://localhost:3000",
		DB:            db,
		authenticator: authenticator,
		users:         map[string]*TestUser{},
		Logger:        logger,
	}

	for _, arg := range os.Args[1:] {
		if arg == "-test.v=true" { // go test transforms -v option
			opts.Format = "pretty"
		}
	}

	var paths []string
	files, err := ioutil.ReadDir(path + "/")
	assert.NoError(t, err)

	paths = make([]string, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".feature") {
			paths = append(paths, fmt.Sprintf("%s/%s", path, f.Name()))
		}
	}

	for _, path := range paths {
		path := path // Pinning ranged variable, more info: https://github.com/kyoh86/scopelint
		t.Run(path, func(t *testing.T) {
			opts.Paths = []string{path}
			suite := godog.TestSuite{
				Name:                 "Integration",
				TestSuiteInitializer: nil,
				ScenarioInitializer:  s.InitializeScenario,
				Options:              &opts,
			}
			status := suite.Run()

			if status != 0 {
				assert.Fail(t, "one or more scenarios failed in feature: "+path)
			}
		})
	}
}
