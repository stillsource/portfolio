Feature: Poem visibility
  Poems appear alongside their photo when the photo is near the viewport center.
  The visibility window should be wide enough for comfortable reading.

  Scenario: Poem is visible when image is centered in viewport
    Given I am on roll page "matin-brumeux"
    When I scroll to image 1
    Then the poem for image 1 should have opacity greater than "0.4"

  Scenario: Poem fades when image is far from viewport center
    Given I am on roll page "matin-brumeux"
    When I scroll past image 1
    Then the poem for image 1 should have opacity less than "0.1"

  Scenario: Typewriter animation types out poem on scroll
    Given I am on roll page "matin-brumeux" with poem animation "typewriter"
    When I scroll to image 1
    Then the poem for image 1 should contain visible text
