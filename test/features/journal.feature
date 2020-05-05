Feature: Journal Tests

  Scenario: 

      #insert an entry
      When we issue a "POST" to "%(domain)s:4001/entry" with payload
      """
      {
        "entry": "hello world"
      }
      """
      Then the response http code is 200
      And we store the response field entry_id

      #get the entry
      When we issue a "GET" to "%(domain)s:4002/entry/%(entry_id)s"
      Then the response http code is 200
      And the response payload resembles
      """
      {
        "entry": "hello world"
      }
      """
