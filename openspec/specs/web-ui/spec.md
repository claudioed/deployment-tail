## ADDED Requirements

### Requirement: Tag-based owner input in web UI

The system SHALL provide a tag-based input component for managing multiple owners with semicolon-separated parsing.

#### Scenario: Display owner tag input on create form
- **WHEN** user accesses schedule create form
- **THEN** system displays owner input field as a tag input component

#### Scenario: Display owner tag input on edit form
- **WHEN** user accesses schedule edit form
- **THEN** system displays owner input field pre-populated with existing owners as tags

#### Scenario: Enter owners separated by semicolon
- **WHEN** user types "john.doe;jane.smith;bob.jones" in owner input
- **THEN** system parses input and creates three separate owner tags

#### Scenario: Create tag on blur
- **WHEN** user types "john.doe" and clicks outside the input field
- **THEN** system creates owner tag for "john.doe"

#### Scenario: Create tag on Enter key
- **WHEN** user types "john.doe" and presses Enter key
- **THEN** system creates owner tag for "john.doe"

#### Scenario: Create tag on semicolon
- **WHEN** user types "john.doe;"
- **THEN** system immediately creates owner tag for "john.doe" and clears input

#### Scenario: Remove owner tag
- **WHEN** user clicks X button on an owner tag
- **THEN** system removes that owner tag from the list

#### Scenario: Display placeholder text
- **WHEN** user views empty owner input
- **THEN** system displays placeholder "Enter owners separated by ;"

#### Scenario: Validate owner format on tag creation
- **WHEN** user enters invalid owner format "invalid@#$"
- **THEN** system displays validation error "Invalid owner format: invalid@#$" and does not create tag

#### Scenario: Prevent duplicate owners
- **WHEN** user enters owner "john.doe" that already exists as a tag
- **THEN** system displays message "Owner already added" and does not create duplicate tag

#### Scenario: Trim whitespace from owner input
- **WHEN** user enters " john.doe "
- **THEN** system creates owner tag with trimmed value "john.doe"

#### Scenario: Visual tag styling
- **WHEN** user views owner tags
- **THEN** each tag displays owner text with X button and distinct visual styling (background color, rounded borders)

#### Scenario: Submit with multiple owners
- **WHEN** user submits form with owner tags ["john.doe", "jane.smith"]
- **THEN** system sends owners as array in API request

### Requirement: Tag-based environment input in web UI

The system SHALL provide a tag-based input component for managing multiple environments.

#### Scenario: Display environment tag input on create form
- **WHEN** user accesses schedule create form
- **THEN** system displays environment selection as a tag interface with predefined options

#### Scenario: Display environment tag input on edit form
- **WHEN** user accesses schedule edit form
- **THEN** system displays environment input pre-populated with existing environments as tags

#### Scenario: Show available environment options
- **WHEN** user focuses on environment input
- **THEN** system displays available environment options (production, staging, development) that are not already selected

#### Scenario: Add environment tag by clicking option
- **WHEN** user clicks "production" from available options
- **THEN** system adds "production" tag to selected environments

#### Scenario: Remove environment tag
- **WHEN** user clicks X button on an environment tag
- **THEN** system removes that environment tag and makes it available for selection again

#### Scenario: Color-coded environment tags
- **WHEN** user views environment tags
- **THEN** production tag displays in red, staging tag in yellow, development tag in green

#### Scenario: Prevent duplicate environments
- **WHEN** user attempts to add environment "production" that is already selected
- **THEN** system does not create duplicate tag

#### Scenario: Validate at least one environment
- **WHEN** user attempts to remove last environment tag
- **THEN** system displays error "At least one environment is required" and prevents removal

#### Scenario: Submit with multiple environments
- **WHEN** user submits form with environment tags ["production", "staging"]
- **THEN** system sends environments as array in API request

### Requirement: Inline status editing in schedule list

The system SHALL allow users to edit deployment schedule status directly from the list view.

#### Scenario: Display status as clickable badge
- **WHEN** user views schedule list
- **THEN** each schedule row displays current status as a clickable badge

#### Scenario: Click to edit status
- **WHEN** user clicks on a status badge in the list
- **THEN** system displays inline dropdown with available status options (created, approved, denied)

#### Scenario: Select new status
- **WHEN** user selects a different status from the inline dropdown
- **THEN** system immediately updates the UI to show the new status

#### Scenario: Show loading state during update
- **WHEN** user selects a new status and API call is in progress
- **THEN** system displays loading spinner on the status badge

#### Scenario: Successful status update
- **WHEN** API call completes successfully
- **THEN** system keeps the updated status displayed and shows success toast message

#### Scenario: Failed status update
- **WHEN** API call returns an error
- **THEN** system reverts the status badge to the previous value and displays error toast message

#### Scenario: Close dropdown on outside click
- **WHEN** user clicks outside the dropdown without selecting a status
- **THEN** system closes the dropdown and keeps the current status unchanged

#### Scenario: Keyboard support for status editing
- **WHEN** user focuses on status badge and presses Enter
- **THEN** system opens the inline dropdown for status selection

### Requirement: Display multiple owners in schedule list

The system SHALL display multiple owners for each schedule in the list view.

#### Scenario: Display owners as comma-separated text
- **WHEN** user views schedule list
- **THEN** each schedule row displays all owners as comma-separated text (e.g., "john.doe, jane.smith")

#### Scenario: Truncate long owner list
- **WHEN** schedule has more than 3 owners
- **THEN** system displays first 3 owners followed by "+N more" text

#### Scenario: Show full owner list on hover
- **WHEN** user hovers over truncated owner text with "+N more"
- **THEN** system displays tooltip with full list of all owners

### Requirement: Display multiple environments in schedule list

The system SHALL display multiple environments for each schedule in the list view.

#### Scenario: Display environment badges
- **WHEN** user views schedule list
- **THEN** each schedule row displays environment badges for all assigned environments

#### Scenario: Color-coded environment badges in list
- **WHEN** user views schedule with environments ["production", "staging"]
- **THEN** system displays production badge in red and staging badge in yellow

#### Scenario: Multiple environment badges side-by-side
- **WHEN** schedule has multiple environments
- **THEN** system displays all environment badges horizontally adjacent to each other

### Requirement: Sidebar-based layout

The system SHALL display a fixed left sidebar with main content area, and the layout SHALL include a header region containing a user identity chip.

#### Scenario: Desktop shows sidebar and content side-by-side
- **WHEN** user accesses application on desktop (viewport >= 768px)
- **THEN** system displays fixed 240px sidebar on left and flexible content area on right

#### Scenario: Sidebar contains group navigation
- **WHEN** user views sidebar
- **THEN** sidebar displays "All Schedules", "Ungrouped", and all accessible groups

#### Scenario: Content area displays schedules
- **WHEN** user views main content area
- **THEN** content displays schedules grouped by date for selected sidebar item

#### Scenario: Sidebar has visual separation
- **WHEN** user views interface
- **THEN** sidebar has distinct background color and border to separate from main content

#### Scenario: Header contains user identity chip
- **WHEN** user views the application header
- **THEN** the header contains a user chip at the top-right showing the authenticated user

### Requirement: Mobile-responsive sidebar

The system SHALL adapt sidebar for mobile viewports.

#### Scenario: Mobile hides sidebar by default
- **WHEN** viewport width is < 768px
- **THEN** sidebar is hidden from view
- **AND** hamburger menu icon appears in header

#### Scenario: User opens mobile sidebar
- **WHEN** user taps hamburger icon on mobile
- **THEN** sidebar slides in from left as overlay
- **AND** backdrop dims main content

#### Scenario: User closes mobile sidebar by backdrop
- **WHEN** user taps backdrop on mobile
- **THEN** sidebar slides out and hides

#### Scenario: User closes mobile sidebar by selection
- **WHEN** user taps a group in mobile sidebar
- **THEN** sidebar automatically slides out
- **AND** main content shows selected group's schedules

### Requirement: Schedule list with date grouping

The system SHALL organize schedule list by relative date sections.

#### Scenario: Display date section headers
- **WHEN** user views schedule list
- **THEN** schedules are grouped under headers: Today, Tomorrow, This Week, Later

#### Scenario: Each section shows count
- **WHEN** user views date section
- **THEN** section header displays count of schedules (e.g., "Today (3)")

#### Scenario: Sections are collapsible
- **WHEN** user clicks date section header
- **THEN** section collapses or expands
- **AND** collapse state persists in localStorage

#### Scenario: Empty sections are hidden
- **WHEN** a date section has zero schedules
- **THEN** that section header does not appear

#### Scenario: Schedules within section show time
- **WHEN** user views schedule in date section
- **THEN** schedule displays time in HH:MM format

### Requirement: Group creation from sidebar

The system SHALL allow users to create groups directly from sidebar.

#### Scenario: Sidebar shows create button
- **WHEN** user views sidebar
- **THEN** "+ New Group" button appears at top of group list

#### Scenario: Click create opens modal
- **WHEN** user clicks "+ New Group" button
- **THEN** group creation modal opens

#### Scenario: Modal includes visibility toggle
- **WHEN** user views group creation modal
- **THEN** modal displays visibility toggle with options: Public, Private

#### Scenario: Create and add to sidebar
- **WHEN** user creates a group
- **THEN** new group appears in sidebar immediately
- **AND** group is positioned based on favorite status and alphabetical order

### Requirement: Inline group management from sidebar

The system SHALL provide quick access to group settings from sidebar.

#### Scenario: Owner sees settings icon on hover
- **WHEN** user hovers over their own group in sidebar
- **THEN** settings gear icon appears next to group name

#### Scenario: Non-owner does not see settings icon
- **WHEN** user hovers over another user's public group
- **THEN** no settings icon appears

#### Scenario: Click settings opens edit modal
- **WHEN** user clicks settings icon
- **THEN** group edit modal opens with current values

#### Scenario: Favorite toggle in sidebar
- **WHEN** user clicks star icon next to group name
- **THEN** group favorite status toggles
- **AND** group repositions in sidebar (favorites first)

### Requirement: Toolbar simplified for sidebar layout

The system SHALL simplify toolbar by moving group actions to sidebar.

#### Scenario: Toolbar shows create schedule button
- **WHEN** user views toolbar
- **THEN** "Create Schedule" button remains in toolbar

#### Scenario: Toolbar shows refresh button
- **WHEN** user views toolbar
- **THEN** "Refresh" button remains in toolbar

#### Scenario: Group buttons removed from toolbar
- **WHEN** user views toolbar
- **THEN** "Create Group" and "Manage Groups" buttons are no longer in toolbar (moved to sidebar)

#### Scenario: Filters remain in toolbar
- **WHEN** user views toolbar
- **THEN** environment and status filter dropdowns remain available

### Requirement: Header displays logged user chip

The system SHALL display a user identity chip in the top-right of the header on every authenticated page.

#### Scenario: Chip is present in header
- **WHEN** an authenticated user views any view of the web UI (list, form, detail)
- **THEN** the header contains a user chip positioned at the top-right

#### Scenario: Chip persists across view changes
- **WHEN** the user navigates between list, form, and detail views
- **THEN** the user chip remains visible and its content does not re-flicker

### Requirement: Header is compacted to accommodate user chip

The system SHALL reduce header vertical padding and inline the subtitle to accommodate the user chip without growing header height.

#### Scenario: Compact header on desktop
- **WHEN** a desktop user (viewport >= 768px) views the application
- **THEN** header vertical padding is reduced compared to the pre-change version
- **AND** the h1 title and the chip fit on the same row

#### Scenario: Subtitle remains readable
- **WHEN** the header is compacted
- **THEN** the subtitle text is still legible and associated with the h1

#### Scenario: Chip collapses on narrow viewports
- **WHEN** the viewport is narrower than 480px
- **THEN** the chip displays only the avatar
- **AND** the header still fits on a single row with the hamburger menu and title

### Requirement: Notification uses accessible live region

The system SHALL expose the notification container as an ARIA live region so messages are announced to assistive technology.

#### Scenario: Notification container attributes
- **WHEN** the notification element is rendered
- **THEN** it has `role="status"` and `aria-live="polite"` and `aria-atomic="true"`

#### Scenario: Error notifications are assertive
- **WHEN** an error notification is shown
- **THEN** the container uses an assertive live region so the message is announced immediately

#### Scenario: Notification does not interrupt unrelated work
- **WHEN** a success notification appears while the user is typing in a form field
- **THEN** the polite live region allows the announcement to queue without stealing focus

### Requirement: Standardized component interactive states

The system SHALL provide consistent visual styles for interactive states on buttons, sidebar items, and user chip controls.

#### Scenario: Hover state on buttons
- **WHEN** a mouse user hovers over a button
- **THEN** the button displays a hover overlay using the `--color-hover-overlay` token

#### Scenario: Focus-visible state on buttons
- **WHEN** a keyboard user focuses a button
- **THEN** the button displays a visible focus ring using `--color-focus-ring`

#### Scenario: Active/pressed state on buttons
- **WHEN** a user presses a button
- **THEN** the button displays a pressed overlay using `--color-pressed-overlay`

#### Scenario: Disabled state on buttons
- **WHEN** a button is disabled
- **THEN** it displays reduced opacity and does not respond to hover, focus, or click

#### Scenario: Loading state on buttons
- **WHEN** a button is in a loading state (e.g., during API call)
- **THEN** it displays a loading indicator
- **AND** is not clickable

### Requirement: currentUser state is the source of truth for user identity

The system SHALL maintain a module-level `currentUser` state in `web/app.js` populated from the authenticated user profile endpoint and read by all code that needs the authenticated user's identifier.

#### Scenario: currentUser populated before data loads
- **WHEN** the application bootstraps
- **THEN** `currentUser` is populated from the profile endpoint before `loadGroupsAndRenderSidebar()` or `loadSchedules()` is called

#### Scenario: getCurrentUser reads from currentUser
- **WHEN** any code calls `getCurrentUser()`
- **THEN** the return value is derived from the `currentUser` state object
- **AND** never from the owner filter input field

#### Scenario: Owner filter is independent of identity
- **WHEN** the user types in the owner filter field
- **THEN** `currentUser` is not modified
- **AND** subsequent API calls that depend on identity still use the authenticated user

## Notes

- Tag input components built with vanilla JavaScript (no framework dependencies)
- Owner tag input supports both manual typing and semicolon parsing
- Environment tag input uses predefined options (production, staging, development only)
- Environment colors: production=red (#ef4444), staging=yellow (#f59e0b), development=green (#10b981)
- Inline status editing uses optimistic UI updates with rollback on error
- Owner and environment validation happens client-side before submission
- At least one owner and one environment required for form submission
- Tag components use consistent styling across create and edit forms
- Tooltips implemented with CSS hover (no JavaScript tooltip library needed)

## Affected Components

- **HTML**: Update `web/index.html` with tag input markup, inline edit elements
- **CSS**: Add `web/styles.css` styles for tags, color-coded badges, tooltips, loading states
- **JavaScript**: Add tag input controllers, inline edit handlers, API integration in `web/app.js`
- **No Backend Changes**: Uses updated schedule API endpoints

## Rollback Plan

1. Remove tag input components from `web/index.html`
2. Restore single-value owner and environment inputs
3. Remove inline edit functionality
4. Remove tag-related styles from `web/styles.css`
5. Remove tag controllers and inline edit handlers from `web/app.js`
6. Revert to dropdown for environment, text input for owner
