Feature: Journal Tests

  Scenario: 

      #insert an journal
      When we issue a "POST" to "%(domain)s:4001/journal" with payload
      """
      {
        "entry": "hello world",
        "created":"2020-05-05T13:03:57Z"
      }
      """
      Then the response http code is 200
      And we store the response field journal_id

      #get the journal
      When we issue a "GET" to "%(domain)s:4002/journal/%(journal_id)s"
      Then the response http code is 200
      And the response payload resembles
      """
      {
        "entry": "hello world",
        "created":"2020-05-05T13:03:57Z"
      }
      """
