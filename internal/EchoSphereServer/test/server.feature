Feature: Server Handling Client Connections and Messages

  Scenario: Server receives message X and forwards it to a randomly selected connection
    Given a server is running and listening for connections
    And a client is connected to the server
    When the client sends "message X"
    Then the server should forward "message X" to a randomly selected connection

  Scenario: Server receives ok X and forwards it to the original sender
    Given a server is running and listening for connections
    And a client is connected to the server
    And the client has previously sent "message X"
    When the server receives "ok X"
    Then the server should forward "ok X" to the original sender of "message X"

  Scenario: Server removes closed connection from active connections
    Given a server is running and listening for connections
    And a client is connected to the server
    When the client disconnects
    Then the server should remove the connection from the active connections