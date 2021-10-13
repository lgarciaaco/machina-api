// Creating a user in a random organization:
//      Given a user named "Bob"
// Creating a user in a given organization:
//      Given a user named "Jimmy" in organization "13639843"
// Logging into a user session:
//      Given I am logged in as "Jimmy"
// Setting the Authorization header of the current user session:
//      Given I set the Authorization header to "Bearer ${agent_token}"
package cucumber

import (
	"context"
	"time"

	"github.com/cucumber/godog"
	"github.com/lgarciaaco/machina-api/business/core/user"
	"github.com/pkg/errors"
)

type contextKey string

var ContextAccessToken = contextKey("accesstoken")

func init() {
	StepModules = append(StepModules, func(ctx *godog.ScenarioContext, s *TestScenario) {
		ctx.Step(`^a user with id "([^"]*)" and password "([^"]*)"$`, s.Suite.verifyUserAndPassword)
		ctx.Step(`^I am logged in as "([^"]*)"$`, s.iAmLoggedInAs)
		ctx.Step(`^I set the "([^"]*)" header to "([^"]*)"$`, s.iSetTheHeaderTo)
	})
}

func (s *TestSuite) verifyUserAndPassword(usrId string, password string) error {
	// users are shared concurrently across scenarios.. so lock while we create the user...
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.users[usrId] != nil {
		return nil
	}

	ctx := context.Background()
	now := time.Now()

	core := user.NewCore(s.Logger, s.Db)
	u, err := core.QueryByID(ctx, usrId)
	if err != nil {
		return errors.Wrap(err, "retrieving user")
	}

	claims, err := core.Authenticate(ctx, now, u.ID, password)
	if err != nil {
		return err
	}

	token, err := s.authenticator.GenerateToken(claims)
	if err != nil {
		return err
	}

	s.users[usrId] = &TestUser{
		Name:  usrId,
		Token: token,
		Ctx:   context.WithValue(context.Background(), ContextAccessToken, token),
	}
	return nil
}

func (s *TestScenario) iAmLoggedInAs(name string) error {
	s.Session().Header.Del("Authorization")
	s.CurrentUser = name
	return nil
}

func (s *TestScenario) iSetTheHeaderTo(name string, value string) error {
	expanded, err := s.Expand(value)
	if err != nil {
		return err
	}

	s.Session().Header.Set(name, expanded)
	return nil
}
