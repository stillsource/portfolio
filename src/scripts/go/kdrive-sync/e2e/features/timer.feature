Feature: Navigation timer
  The footer timer auto-navigates to the next roll after 4 seconds.
  If the user scrolls back up, the timer must reset immediately to 0.

  Scenario: Timer resets to 0 when user scrolls back up
    Given I am on roll page "matin-brumeux"
    When I scroll to the bottom of the page
    Then the navigation timer should start
    When I scroll back to the top
    Then the progress bar value should be "0"
    And the navigation timer should be cancelled

  Scenario: Timer completes and navigates after 4 seconds
    Given I am on roll page "matin-brumeux"
    When I scroll to the bottom of the page
    Then the navigation timer should start
    And after 4500 milliseconds the page should have navigated
