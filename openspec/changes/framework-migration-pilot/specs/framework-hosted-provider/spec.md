## Purpose

Lets the provider host one or more resources on `terraform-plugin-framework` while continuing to host other resources on `terraform-plugin-sdk/v2`, without invalidating existing Terraform state files for resources that have not yet been migrated.

## ADDED Requirements

### Requirement: Provider can host resources on Plugin Framework

The provider SHALL be able to host one or more resources on `terraform-plugin-framework` while continuing to host other resources on `terraform-plugin-sdk/v2`. Both runtimes MUST initialize during provider startup.

#### Scenario: Mixed-runtime provider initializes
- **WHEN** a Terraform configuration references both a framework-hosted resource and an SDK v2 resource
- **THEN** the provider initializes both runtimes and `terraform init` completes without error

### Requirement: State file compatibility preserved across migration

The provider MUST NOT invalidate existing Terraform state files when a resource is migrated from SDK v2 to Plugin Framework. A state file written before migration SHALL plan unchanged against the framework implementation. Resource addresses, attribute names, and attribute types SHALL remain stable across the migration.

#### Scenario: SDK v2 state plans unchanged against framework implementation
- **WHEN** a state file contains a resource written by the SDK v2 implementation
- **THEN** `terraform plan` against the framework implementation produces no diff and no replacement

#### Scenario: Resource address and attribute names remain stable
- **WHEN** a resource is migrated from SDK v2 to Plugin Framework
- **THEN** its state address, attribute names, and attribute types are unchanged from the SDK v2 implementation

#### Scenario: State plans unchanged across namespace handoff
- **WHEN** a state file written under the legacy `kislerdm/neon` provider source is planned against a configuration using the new `neon/neon` source for the same resources
- **THEN** `terraform plan` is empty (no diff, no replacement), because state files reference resource addresses rather than provider sources

### Requirement: Framework-hosted resources tolerate out-of-band deletion

Framework-hosted resources SHALL tolerate deletion that occurs outside Terraform (for example, deletion via the Neon console) without leaving the provider in an unrecoverable state. The FSM pattern (`ReadRetry`, `DeleteRetry`, `Import`) used by the SDK v2 implementation SHALL be available to framework-hosted resources.

#### Scenario: Refresh after out-of-band deletion removes resource from state
- **WHEN** a framework-hosted resource is deleted outside Terraform and `terraform plan` is run
- **THEN** the resource is removed from state with no diff and no error

#### Scenario: Destroy is tolerant when the resource no longer exists
- **WHEN** `terraform destroy` is invoked and the framework-hosted resource no longer exists in the Neon project
- **THEN** the destroy completes without error

### Requirement: Acceptance tests cover refresh, destroy, update, import

For each framework-hosted resource, acceptance tests SHALL exercise four cases: refresh, destroy, update, and import. All four cases MUST pass against a real Neon project before the framework-hosted resource is considered ready for release.

#### Scenario: All four test cases pass for the pilot resource
- **WHEN** the acceptance test suite runs against the framework-hosted pilot resource
- **THEN** refresh, destroy, update, and import cases all pass
