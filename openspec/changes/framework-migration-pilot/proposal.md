## Why

The provider runs on `terraform-plugin-sdk/v2`, which HashiCorp has signaled for deprecation. The strategic plan in `openspec/specs/original-spec.md` proposes a phased migration to `terraform-plugin-framework`, with Phase 2 being a pilot on one resource. Main has since moved to v0.15.0 (commits `8ddaf91`, `2803f85`), which added FSM coverage to more resources (JWKS, API keys). That makes the migration template more consequential than the original plan anticipated — a pilot that omits FSM handlers would mislead the next nine resources. We need to validate the migration pattern on a real resource before committing to bulk migration.

## What Changes

- Port `neon_api_key` to `terraform-plugin-framework` as the migration pilot
- Include FSM handlers (`ReadRetry`, `DeleteRetry`, `Import`) in the pilot so it serves as a template for resources that need out-of-band deletion tolerance
- Rebase the existing `feat/migrate-sdk-v2-to-framework` branch onto `main` (`c81958f`) before landing — that branch is now three commits behind and predates the v0.15.0 sync
- Preserve state file compatibility: a state file written by the SDK v2 implementation must plan unchanged against the framework implementation, with no forced re-import
- Acceptance tests cover refresh, destroy, update, and import — matching the FSM test structure already established on main
- Release per the original-spec recommendation as part of the v1.0.0 transition; this change alone does not trigger the major bump, but lands as a pre-release tag so users can opt in
- This change is the **start** of the SDK v2 -> Plugin Framework migration path. The two-namespace state (`kislerdm/neon` and `neon/neon`) is transitional; the destination is a single `neon/neon` namespace once Phase 0 governance work (namespace redirect or transfer with the original maintainer) and Phase 3 bulk resource migration complete

No **BREAKING** changes. State compatibility is the explicit promise of this change.

## Capabilities

### New Capabilities

- `framework-hosted-provider`: the provider can host one or more resources on `terraform-plugin-framework` while preserving state-file compatibility for resources that have not yet been migrated. The first such resource is `neon_api_key`.

### Modified Capabilities

None. `openspec/specs/` currently holds only the strategic plan (`original-spec.md`), which is not a formal capability spec under this schema.

## Impact

- `provider/resource_api_key.go` — rewritten on Framework
- `provider/acc_api_key_test.go` — acceptance tests preserved; framework wiring added
- `provider/retry.go` — FSM handlers reviewed for Framework compatibility and adapted if needed
- `go.mod` — Framework dependency added alongside existing SDK v2 (no removal yet)
- Build artifact — published as a pre-release tag; not the v1.0.0 bump itself
- `feat/migrate-sdk-v2-to-framework` branch — rebased onto `c81958f` as the first task

## Out of Scope

- Migrating any resource other than `neon_api_key` — separate changes
- Extracting the migration template into reusable helpers — Change 2
- Adding new resources beyond the existing surface — Phase 1 in the strategic plan
- Resolving governance decisions in the strategic plan (namespace, signing keys, maintainer model) — Phase 0
