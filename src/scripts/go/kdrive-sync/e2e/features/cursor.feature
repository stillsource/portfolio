Feature: Custom cursor
  The custom cursor dot must remain visible after page navigations
  and reflect hover/reading states correctly.

  Scenario: Cursor stays visible after navigating between rolls
    Given I am on the homepage
    When I move the mouse to position 400,300
    Then the cursor should be visible
    When I navigate to roll "matin-brumeux"
    And I move the mouse to position 400,400
    Then the cursor should still be visible

  Scenario: Cursor disappears when mouse leaves the browser window
    Given I am on the homepage
    When I move the mouse to position 400,300
    Then the cursor should be visible
    When the mouse leaves the browser window
    Then the cursor should not be visible
