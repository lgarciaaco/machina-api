Feature: Users api
  User represents someone with access to our system. Users can have
  the role 'admin' or 'user'. Users with role 'user' can only access resources
  created by themselves.

  Background:
    Given the path prefix is "/v1"
    Given a user with id "5cf37266-3473-4006-984f-9325122678b7" and password "gophers"
    Given a user with id "45b5fbd3-755f-4379-8f07-a58d4a30fa2f" and password "gophers"

  Scenario: a user must provide user_id and password to get a token
    When I GET path "/users/token"
    Then the response code should be 401
    And the response should match json:
    """
    {"error":"must provide id and password in Basic auth"}
    """

  Scenario: can't create a user unless authenticated
    When I POST path "/users" with json body:
    """
    {"name": "some_name", "description": "desc", "password": "some_password", "roles": "ADMIN"}
    """
    Then the response code should be 401
    And the response should match json:
    """
    {"error": "expected authorization header format: Bearer: TOKEN"}
    """

  Scenario: can't create a user unless you are an admin
    Given I am logged in as "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
    When I POST path "/users" with json body:
    """
    {"name": "some_name", "description": "desc", "password": "some_password", "roles": "ADMIN"}
    """
    Then the response code should be 403
    And the response should match json:
    """
    {"error": "you are not authorized for that action, claims[[USER]] roles[[ADMIN]]"}
    """

  Scenario: can't create a user with a malformed body
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I POST path "/users" with json body:
    """
    {"name": "some_name"}
    """
    Then the response code should be 400
    And the response should match json:
    """
    {
      "error":"data validation error",
      "fields":{
        "password":"password is a required field",
        "description": "description is a required field",
        "roles":"roles is a required field"
      }
    }
    """

  Scenario: when fetching a user, the user id should be valid
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I GET path "/users/some_bad_id"
    Then the response code should be 400
    And the response should match json:
    """
    {"error": "ID is not in its proper form"}
    """

  Scenario: a user with admin role request for a user that does not exist within the endpoint
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I GET path "/users/c50a5d66-3c4d-453f-af3f-bc960ed1a503"
    Then the response code should be 404
    And the response should match json:
    """
    {"error": "user not found"}
    """

  Scenario: a users with admin role can page through all users
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I GET path "/users/1/2"
    Then the response code should be 200
    And the response should match json:
    """
    [
      {
         "date_created": "2019-03-26T00:00:00Z",
         "date_updated": "2019-03-24T00:00:00Z",
         "description": "This strategy triggers when the price crosses above or below a moving average signalising a potential change in trend. It uses two moving averages to find trend and opportunity.",    "id": "5af36f98-b5d2-4991-b2ba-95442550897b",
         "name": "SMA BNBUSDT 4H",
         "positions_total": 0,
         "roles": [
            "USER"
         ]
      },
       {
         "date_created": "2019-03-25T00:00:00Z",
         "date_updated": "2019-03-24T00:00:00Z",
         "description": "This strategy triggers when the price crosses above or below a moving average signalising a potential change in trend. It uses two moving averages to find trend and opportunity.",
         "id": "d29b4b23-c003-4519-b3af-051b9c9b3c5a",
         "name": "SMA BNBUSDT 1H",
         "positions_total": 0,
         "roles": [
            "USER"
         ]
      }
    ]
    """

  Scenario: a regular user can only fetch themselves.
    Given I am logged in as "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
    When I GET path "/users/1/1"
    Then the response code should be 403
    And the response should match json:
    """
    {"error": "you are not authorized for that action, claims[[USER]] roles[[ADMIN]]"}
    """

    When I GET path "/users/5cf37266-3473-4006-984f-9325122678b7"
    Then the response code should be 403
    And the response should match json:
    """
    {"error": "attempted action is not allowed"}
    """

    When I DELETE path "/users/5cf37266-3473-4006-984f-9325122678b7"
    Then the response code should be 403
    And the response should match json:
    """
    {"error": "you are not authorized for that action, claims[[USER]] roles[[ADMIN]]"}
    """

    When I GET path "/users/45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
    Then the response code should be 200
    And the response should match json:
    """
    {
     "date_created": "2019-03-24T00:00:00Z",
     "date_updated": "2019-03-24T00:00:00Z",
     "id": "45b5fbd3-755f-4379-8f07-a58d4a30fa2f",
     "name": "User Gopher",
     "description": "This is some regular user",
     "positions_total": 3,
     "roles": [
        "USER"
     ]
    }
    """

  Scenario: deleting a user that does not exist is not a failure
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I DELETE path "/users/5cf37266-3473-4006-984f-932512267357"
    Then the response code should be 204

  Scenario: an admin can do CRUD on a user
    Given I am logged in as "5cf37266-3473-4006-984f-9325122678b7"
    When I POST path "/users" with json body:
    """
    {"name": "some_name", "description": "desc", "password": "some_password", "password_confirm": "some_password", "roles": ["ADMIN"]}
    """
    Then the response code should be 201
    Given I store the ".id" selection from the response as ${nusrID}
    And the response should match json:
    """
    {
     "date_created": "${response.date_created}",
     "date_updated": "${response.date_updated}",
     "id": "${response.id}",
     "name": "some_name",
     "description": "desc",
     "positions_total": 0,
     "roles": [
        "ADMIN"
     ]
    }
    """

    When I GET path "/users/${nusrID}"
    Then the response code should be 200
    And the response should match json:
    """
    {
     "date_created": "${response.date_created}",
     "date_updated": "${response.date_updated}",
     "id": "${response.id}",
     "name": "some_name",
     "description": "desc",
     "positions_total": 0,
     "roles": [
        "ADMIN"
     ]
    }
    """

    When I PUT path "/users/${nusrID}" with json body:
    """
    {"name": "some_other_name"}
    """
    Then the response code should be 204

    When I DELETE path "/users/${nusrID}"
    Then the response code should be 204