package cucumber

import (
	"context"
	"time"

	"github.com/cucumber/godog"
	"github.com/lgarciaaco/machina-api/business/core/user"
	"github.com/pkg/errors"
)

type contextKey string

var contextAccessToken = contextKey("accesstoken")

func init() {
	StepModules = append(StepModules, func(ctx *godog.ScenarioContext, s *TestScenario) {
		ctx.Step(`^a user with id "([^"]*)" and password "([^"]*)"$`, s.Suite.verifyUserAndPassword)
		ctx.Step(`^I am logged in as "([^"]*)"$`, s.iAmLoggedInAs)
		ctx.Step(`^I set the "([^"]*)" header to "([^"]*)"$`, s.iSetTheHeaderTo)
	})
}

func (s *TestSuite) verifyUserAndPassword(usrID string, password string) error {
	// users are shared concurrently across scenarios.. so lock while we create the user...
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.users[usrID] != nil {
		return nil
	}

	ctx := context.Background()
	now := time.Now()

	core := user.NewCore(s.Logger, s.DB)
	u, err := core.QueryByID(ctx, usrID)
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

	s.users[usrID] = &TestUser{
		Name:  usrID,
		Token: token,
		Ctx:   context.WithValue(context.Background(), contextAccessToken, token),
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
