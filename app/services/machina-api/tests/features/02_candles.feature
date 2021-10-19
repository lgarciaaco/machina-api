Feature: Candles api
  Candles are the core unit of strategies. Candles api
  doesnt require authentication and it fetches candles

  Background:
    Given the path prefix is "/v1"

  Scenario: List / Fetch candles, paged, newest candles should come first
    When I GET path "/candles/1/3"
    Then the response code should be 200
    And the response should match json:
    """
    [
      {
        "close_price": 310.5,
        "close_time": "2019-01-01T00:09:03.000001Z",
        "high": 311,
        "id": "cd0f4919-2fe7-4808-8ba2-a1ea652cd591",
        "interval": "4h",
        "low": 398.6,
        "open_price": 200.5,
        "open_time": "2019-01-01T00:08:03.000001Z",
        "symbol": "ETHUSDT",
        "volume": 33456
      },
      {
        "close_price": 210.5,
        "close_time": "2019-01-01T00:08:03.000001Z",
        "high": 211,
        "id": "01d92444-71a7-450f-8a1b-e488cb1a6973",
        "interval": "4h",
        "low": 198.6,
        "open_price": 200.5,
        "open_time": "2019-01-01T00:04:03.000001Z",
        "symbol": "ETHUSDT",
        "volume": 23456
      },
      {
        "close_price": 110.5,
        "close_time": "2019-01-01T00:04:03.000001Z",
        "high": 111,
        "id": "039eee35-7463-4dbd-ae91-0428f3b89c42",
        "interval": "4h",
        "low": 98.6,
        "open_price": 100.5,
        "open_time": "2019-01-01T00:00:03.000001Z",
        "symbol": "ETHUSDT",
        "volume": 13456
      }
    ]
    """

    When I GET path "/candles/125240c0-7f7f-4d0f-b30d-939fd93cf027/4h/1/1"
    Then the response code should be 200
    And the response should match json:
    """
    [
      {
        "close_price": 310.5,
        "close_time": "2019-01-01T00:09:03.000001Z",
        "high": 311,
        "id": "cd0f4919-2fe7-4808-8ba2-a1ea652cd591",
        "interval": "4h",
        "low": 398.6,
        "open_price": 200.5,
        "open_time": "2019-01-01T00:08:03.000001Z",
        "symbol": "ETHUSDT",
        "volume": 33456
      }
    ]
    """

    When I GET path "/candles/cd0f4919-2fe7-4808-8ba2-a1ea652cd591"
    Then the response code should be 200
    And the response should match json:
    """
    {
      "close_price": 310.5,
      "close_time": "2019-01-01T00:09:03.000001Z",
      "high": 311,
      "id": "cd0f4919-2fe7-4808-8ba2-a1ea652cd591",
      "interval": "4h",
      "low": 398.6,
      "open_price": 200.5,
      "open_time": "2019-01-01T00:08:03.000001Z",
      "symbol": "ETHUSDT",
      "volume": 33456
    }
    """

  Scenario: use wrong symbol and interval to list candles
    When I GET path "/candles/some_symbol/4h/1/1"
    Then the response code should be 400
    And the response should match json:
    """
    {
      "error": "ID is not in its proper form"
    }
    """

    When I GET path "/candles/125240c0-7f7f-4d0f-b30d-939fd93cf027/some_interval/1/1"
    Then the response code should be 200
    And the response should match json:
    """
    []
    """
