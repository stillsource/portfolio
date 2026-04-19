Feature: Photo layout and dimensions
  Photos can be displayed in different sizes and composed in side-by-side pairs.

  Scenario: Image pair renders side by side on desktop
    Given I am on a roll page with a pair layout
    Then the two paired images should appear in the same horizontal row
    And each image should occupy roughly half the viewport width

  Scenario: Image pair stacks vertically on mobile
    Given I am on a roll page with a pair layout
    And the viewport is 375x812
    Then the two paired images should appear stacked vertically

  Scenario: Size "medium" renders narrower than "full"
    Given I am on a roll page with images of different sizes
    Then the "medium" image should be narrower than the "full" image
