package cucumber

import (
	"bytes"
	"context"
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cucumber/godog"
)

func init() {
	StepModules = append(StepModules, func(ctx *godog.ScenarioContext, s *TestScenario) {
		ctx.Step(`^the path prefix is "([^"]*)"$`, s.theAPIPrefixIs)
		ctx.Step(`^I (GET|POST|PUT|DELETE|PATCH|OPTION) path "([^"]*)"$`, s.sendHTTPRequest)
		ctx.Step(`^I (GET|POST|PUT|DELETE|PATCH|OPTION) path "([^"]*)" as a json event stream$`, s.sendHTTPRequestAsEventStream)
		ctx.Step(`^I (GET|POST|PUT|DELETE|PATCH|OPTION) path "([^"]*)" with json body:$`, s.SendHTTPRequestWithJSONBody)
		ctx.Step(`^I wait up to "([^"]*)" seconds for a GET on path "([^"]*)" response "([^"]*)" selection to match "([^"]*)"$`, s.iWaitUpToSecondsForAGETOnPathResponseSelectionToMatch)
		ctx.Step(`^I wait up to "([^"]*)" seconds for a GET on path "([^"]*)" response code to match "([^"]*)"$`, s.iWaitUpToSecondsForAGETOnPathResponseCodeToMatch)
		ctx.Step(`^I wait up to "([^"]*)" seconds for a response event$`, s.iWaitUpToSecondsForAResponseJSONEvent)
	})
}

func (s *TestScenario) theAPIPrefixIs(prefix string) error {
	s.PathPrefix = prefix
	return nil
}

func (s *TestScenario) sendHTTPRequest(method, path string) error {
	return s.SendHTTPRequestWithJSONBody(method, path, nil)
}

func (s *TestScenario) sendHTTPRequestAsEventStream(method, path string) error {
	return s.SendHTTPRequestWithJSONBodyAndStyle(method, path, nil, true, true)
}

func (s *TestScenario) SendHTTPRequestWithJSONBody(method, path string, jsonTxt *godog.DocString) (err error) {
	return s.SendHTTPRequestWithJSONBodyAndStyle(method, path, jsonTxt, false, true)
}

func (s *TestScenario) SendHTTPRequestWithJSONBodyAndStyle(method, path string, jsonTxt *godog.DocString, eventStream bool, expandJSON bool) (err error) {
	// handle panic
	defer func() {
		switch t := recover().(type) {
		case string:
			err = fmt.Errorf(t)
		case error:
			err = t
		}
	}()

	session := s.Session()

	body := &bytes.Buffer{}
	if jsonTxt != nil {
		expanded := jsonTxt.Content
		if expandJSON {
			expanded, err = s.Expand(expanded)
			if err != nil {
				return err
			}
		}
		body.WriteString(expanded)
	}
	expandedPath, err := s.Expand(path)
	if err != nil {
		return err
	}
	fullURL := s.Suite.APIURL + s.PathPrefix + expandedPath

	// Lets reset all the response session state...
	if session.Resp != nil {
		_ = session.Resp.Body.Close()
	}
	session.EventStream = false
	session.Resp = nil
	session.RespBytes = nil
	session.respJSON = nil

	ctx := session.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return err
	}

	// We consume the session headers on every request except for the Authorization header.
	req.Header = session.Header
	session.Header = http.Header{}

	if req.Header.Get("Authorization") != "" {
		session.Header.Set("Authorization", req.Header.Get("Authorization"))
	} else if session.TestUser != nil && session.TestUser.Token != "" {
		req.Header.Set("Authorization", "Bearer "+session.TestUser.Token)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := session.Client.Do(req)
	if err != nil {
		return err
	}

	session.Resp = resp
	session.EventStream = eventStream
	if !eventStream {
		defer resp.Body.Close()
		session.RespBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	} else {

		c := make(chan interface{})
		session.EventStreamEvents = c
		go func() {
			d := json.NewDecoder(session.Resp.Body)
			for {
				var event interface{}
				err := d.Decode(&event)
				if err != nil {
					close(c)
					return
				}
				c <- event
			}
		}()
	}

	return nil
}

func (s *TestScenario) iWaitUpToSecondsForAResponseJSONEvent(timeout float64) error {
	session := s.Session()
	if !session.EventStream {
		return fmt.Errorf("the last http request was not performed as a json event stream")
	}

	session.respJSON = nil
	session.RespBytes = session.RespBytes[0:0]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	defer cancel()

	select {
	case event := <-session.EventStreamEvents:

		session.respJSON = event
		var err error
		session.RespBytes, err = json.Marshal(event)
		if err != nil {
			return err
		}
	case <-ctx.Done():
	}

	return nil
}

func (s *TestScenario) iWaitUpToSecondsForAGETOnPathResponseCodeToMatch(timeout float64, path string, expected int) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	defer cancel()
	session := s.Session()
	session.Ctx = ctx
	defer func() {
		session.Ctx = nil
	}()

	for {
		err := s.sendHTTPRequest("GET", path)
		if err == nil {
			err = s.theResponseCodeShouldBe(expected)
			if err == nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(time.Duration(timeout * float64(time.Second) / 10.0))
		}
	}
}

func (s *TestScenario) iWaitUpToSecondsForAGETOnPathResponseSelectionToMatch(timeout float64, path string, selection, expected string) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	defer cancel()
	session := s.Session()
	session.Ctx = ctx
	defer func() {
		session.Ctx = nil
	}()

	for {
		err := s.sendHTTPRequest("GET", path)
		if err == nil {
			err = s.theSelectionFromTheResponseShouldMatch(selection, expected)
			if err == nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(time.Duration(timeout * float64(time.Second) / 10.0))
		}
	}
}
