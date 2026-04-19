Feature: Aura color system
  The ambient background colors update as the user scrolls through photos
  and when navigating between rolls.

  Scenario: Aura updates when navigating to a new roll
    Given I am on roll page "matin-brumeux"
    And I note the current background color
    When I navigate to roll "nuit-a-tokyo"
    Then the background color should have changed

  Scenario: Aura color does not stay orange for "Vitrine de café" photo
    Given I am on roll page "reflets-de-pluie"
    When I scroll to the photo "Vitrine de café"
    Then the CSS variable "--p1" should not be a warm orange color
