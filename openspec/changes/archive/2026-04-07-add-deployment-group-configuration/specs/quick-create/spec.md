## ADDED Requirements

### Requirement: Group assignment during Quick Create

The system SHALL allow users to optionally assign schedules to groups during Quick Create.

#### Scenario: Create with groups
- **WHEN** user creates schedule via Quick Create and selects groups
- **THEN** system creates schedule and assigns to selected groups

#### Scenario: Groups remain optional
- **WHEN** user creates schedule via Quick Create without selecting groups
- **THEN** system creates schedule successfully without group assignments

#### Scenario: Quick schedule with groups (CLI)
- **WHEN** user runs command with `--groups` flag
- **THEN** system creates schedule and assigns to specified groups

#### Scenario: Groups flag accepts IDs or names
- **WHEN** user provides `--groups "id1,id2"` or `--groups "Project Alpha,Team Backend"`
- **THEN** system resolves groups and assigns schedule accordingly
