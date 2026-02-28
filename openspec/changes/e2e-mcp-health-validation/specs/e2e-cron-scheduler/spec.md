## ADDED Requirements

### Requirement: Schedule lifecycle management
The test suite SHALL verify that the cron scheduler correctly manages schedule registration and removal.

#### Scenario: Add schedule registers cron entry
- **WHEN** AddSchedule is called with a valid cron expression and workflow reference
- **THEN** the scheduler SHALL have an active entry for that workflow and the entry count SHALL increase by 1

#### Scenario: Remove schedule deregisters cron entry
- **WHEN** RemoveSchedule is called for a previously registered workflow
- **THEN** the scheduler SHALL no longer have an entry for that workflow and the entry count SHALL decrease by 1

#### Scenario: Duplicate schedule replaces existing entry
- **WHEN** AddSchedule is called twice for the same workflow with different cron expressions
- **THEN** the scheduler SHALL have exactly one entry for that workflow using the new cron expression

### Requirement: Annotation-based schedule discovery
The test suite SHALL verify that the scheduler discovers cron schedules from deployment annotations.

#### Scenario: Discover schedules from annotated deployments
- **WHEN** DiscoverSchedules is called and the namespace contains deployments with the tentacular.dev/cron-schedule annotation
- **THEN** the scheduler SHALL register entries for each annotated deployment with the correct cron expression

#### Scenario: Skip deployments without cron annotation
- **WHEN** DiscoverSchedules is called and the namespace contains deployments without the cron-schedule annotation
- **THEN** the scheduler SHALL not register entries for those deployments

#### Scenario: Remove stale schedules on re-discover
- **WHEN** DiscoverSchedules is called and a previously discovered deployment no longer has the cron-schedule annotation
- **THEN** the scheduler SHALL remove the stale entry for that deployment

### Requirement: Trigger execution via HTTP
The test suite SHALL verify that a scheduled trigger executes the workflow via HTTP POST.

#### Scenario: Cron trigger calls workflow /run endpoint
- **WHEN** a cron schedule fires (triggered manually in test)
- **THEN** the scheduler SHALL send an HTTP POST to http://<name>.<namespace>.svc.cluster.local:8080/run
