Feature: Client-Server Communication Protocol

  Scenario: Client connects to the server and sends message X
    Given a server is running and listening for connections
    When the client connects to the server
    And the client sends "message X"
    Then the client should listen for responses from the server

  Scenario: Client receives ok X and closes the connection
    Given a client is connected to the server
    And the client has sent "message X"
    When the client receives "ok X"
    Then the client should close the connection

  Scenario: Client receives message Y and replies with ok Y
    Given a client is connected to the server
    When the client receives "message Y"
    Then the client should reply with "ok Y"

  Scenario: Client resends message X if not acknowledged within timeout
    Given a client is connected to the server
    And the client has sent "message X"
    When the client does not receive "ok X" within the timeout period
    Then the client should resend "message X"
    And this process should repeat until "ok X" is received