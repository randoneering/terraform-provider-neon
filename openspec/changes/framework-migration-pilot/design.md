## Context

The provider runs on `terraform-plugin-sdk/v2`. Main is at `c81958f` (v0.15.0 + role 404 fix). A prior pilot attempt lives on `feat/migrate-sdk-v2-to-framework` (tip `9db56df`) and is three commits behind main. The v0.15.0 sync added FSM coverage to `resource_jwks_url.go` and tightened import handling on `resource_api_key.go`, which raises the bar for what the migration template must demonstrate. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**
- Land `neon_api_key` on Plugin Framework with FSM handlers wired in
- Keep the existing pilot branch as the substrate; do not start over
- Validate state compatibility with a CI-runnable test, not a manual check
- Publish as a pre-release tag; do not bump to v1.0.0 yet

**Non-Goals:**
- Migrate any resource other than `neon_api_key`
- Extract the migration template into reusable helpers (Change 2)
- Resolve Phase 0 governance decisions (namespace, signing keys, maintainer model)
- Drop `terraform-plugin-sdk/v2` from `go.mod`

## Decisions

### Adopt the existing pilot branch via rebase, not a fresh start

The branch `feat/migrate-sdk-v2-to-framework` already contains a partial pilot (`b1440f9 pilot neon_api_key`), SDK bump, namespace rename scaffolding, and migration docs. A rebase onto `c81958f` preserves this work and keeps the diff focused on the migration rather than on reproducing baseline state.
- **Alternative A — start fresh from `main`:** discards working pilot code, slower, no upside.
- **Alternative B — cherry-pick only `b1440f9`:** risks missing the SDK bump and rename scaffolding the pilot depends on.

### Include FSM handlers in the pilot, even though `neon_api_key` already has them on SDK v2

The original plan called `neon_api_key` the pilot because it is flat-structured and has no nested state. Since the v0.15.0 sync landed, FSM is a load-bearing pattern that every resource uses (or will use). A pilot that omits the FSM wiring would mislead the next nine migrations. Porting the handlers now makes the pilot honest as a template.
- **Alternative — omit FSM, document as a follow-up:** produces a misleading template; the next change would re-port the same handlers anyway.

### Ship the Framework port as the destination binary under a transitional second namespace

The Framework provider lives in `internal/provider/` and is served by a separate binary built from `cmd/neon-framework/`. It registers under `registry.terraform.io/neon/neon` while the SDK v2 binary (built from the package root) continues to serve `registry.terraform.io/kislerdm/neon`. Both binaries ship side by side during the migration window. Users opt in by setting `source = "neon/neon"` in their `required_providers` block.

This is the **destination** binary for the migration. Once Phase 0 governance work lands (namespace redirect or transfer with the original maintainer) and Phase 3 bulk resource migration completes, only `neon/neon` will remain. The two-namespace state is transitional; it is not a permanent product split.
- **Alternative — muxed single binary (one provider serving both runtimes):** would require `tf6muxserver` to combine SDK v2 and Framework provider servers into one gRPC plugin. Adds multiplexer complexity, harder to ship incrementally, and forces users to migrate before they can opt in.
- **Alternative — replace SDK v2 in-place:** forces every consumer to migrate at once, defeats the incremental release strategy.
- **Alternative — wait for Phase 0 to complete before any Framework binary ships:** serializes the migration plan on a governance decision outside our control. The user-facing migration ergonomics (a one-line `source = ...` change in `required_providers`) work either way, so shipping the destination binary now costs little and unblocks Phase 3.

### Verify state compatibility with an acceptance test, not a manual check

The state-compat promise is the load-bearing constraint of the entire migration. A test that applies with SDK v2 and plans against Framework, asserting zero diff, makes the promise enforceable in CI and reviewable in PRs.
- **Alternative — manual verification only:** fails the moment a contributor changes something without checking.

### Pre-release tag, not the v1.0.0 bump

Per the original-spec decision `T4`, `v1.0.0` should signal the *completed* transition, not the start. Publishing the pilot as a pre-release lets users opt in without the provider claiming the migration is done.
- **Alternative — bump to `v1.0.0` now:** over-promises the state of the codebase to every consumer.

## Risks / Trade-offs

- **FSM handler translation risk.** `retry.go` is written against the SDK v2 state-modification model. Framework's `State` API differs. The handlers may need rewriting rather than direct porting. → Mitigation: review during the port; if the rewrite is non-trivial, defer to a small follow-up change rather than blocking the pilot.
- **Rebase conflicts.** The pilot branch is three commits behind and the v0.15.0 sync touched `resource_api_key.go`. → Mitigation: small, reviewable conflicts; resolve during the rebase task and call out anything beyond mechanical resolution.
- **State-compat test passes only on the happy path.** A test that applies and plans may not catch attribute-name drift or type changes that only show up on real state files. → Mitigation: include out-of-band deletion scenarios and run the test against an existing v0.15.0 state file when feasible.
- **Two binaries to ship, sign, and publish.** The pre-release publishes two artifacts per platform instead of one. Goreleaser config needs a second `builds` entry. → Mitigation: add a Framework build entry to `.goreleaser.yml`; keep the SDK v2 build as the default until v1.0.0.
- **Two binaries to keep compatible on schema.** Each release of the SDK v2 binary and the Framework binary must agree on the resource schema for `neon_api_key` (attribute names, types, ForceNew rules). Drift between them shows up as a state-compat failure. → Mitigation: the schema-diff test in task 4.1 catches drift; both schemas live in the same repo so review covers both.

## Migration Plan

- **Pre-merge:** rebase `feat/migrate-sdk-v2-to-framework` onto `c81958f`; resolve conflicts.
- **Merge:** land via PR after acceptance tests and state-compat test pass.
- **Rollback:** revert the merge commit. State files remain compatible, so no state-level rollback is needed.
- **Post-merge:** publish a pre-release tag; update `CHANGELOG.md` to flag the framework pilot.

## Migration Handoff

The two-namespace state is transitional, not permanent. The destination is a single `neon/neon` namespace once both Phase 0 (governance handover with the original maintainer) and Phase 3 (bulk migration of the remaining 10 resources + 5 data sources to Plugin Framework) complete.

When handoff becomes feasible:

- **All resources on Framework.** `kislerdm/neon` no longer needs to ship, and `neon/neon` serves the full surface (11 resources + 5 data sources).
- **Cross-namespace state compatibility is verified.** A state file written under `kislerdm/neon` plans unchanged against a configuration using `neon/neon` for every resource. State files reference resource addresses, not provider sources, so switching `source = "..."` is a one-line config change for users.
- **Phase 0 governance resolved.** The original maintainer has either transferred the namespace or registered a Terraform Registry redirect from `kislerdm/neon` to `neon/neon`. Deprecation notice is published in the v0.x release line.

This change does not block on handoff. It stands on its own as a pre-release pilot. Handoff is the trigger that lets users start the namespace migration in earnest; the technical capability to do so (state-compat across binaries) is verified as part of this change and the per-resource Phase 3 changes that follow.

## Open Questions

- Whether the FSM helpers end up duplicated between `provider/retry.go` and `internal/provider/retry.go` or get unified behind a thin abstraction. Both copies exist now; unification is a Change 2 candidate once the template stabilizes.
- **Resolved (implementation, transitional):** the framework implementation ships as the destination binary under `neon/neon`; the SDK v2 implementation continues under `kislerdm/neon` for now. Both namespaces are expected to collapse to `neon/neon` after Phase 0 + Phase 3. See the "Ship the Framework port as the destination binary" decision above and the Migration Handoff subsection below.

## OpenSpec Artifact Persistence

The `openspec/` directory is gitignored at the repo root (project policy: planning artifacts stay out of git history). For this change, the artifacts under `openspec/changes/framework-migration-pilot/` were force-added to a dedicated `openspec/framework-migration-pilot` branch on the fork so they travel with the working tree across machines while staying out of feature-branch PRs against the original repo.

Pattern for any change in this repo:

- Push openspec/ artifacts to a dedicated `openspec/<change-name>` branch on the fork.
- Use `git add -f openspec/` to override `.gitignore` on that branch only.
- Feature branches (`feat/...`) stay openspec-free and are safe for upstream PRs.
- On any machine: `git fetch origin openspec/<change-name>` and `git checkout openspec/<change-name>` to recover the planning layer.
- Capture cross-machine resume prompts at `openspec/handoff/<date>-<change>.md` using the same force-add pattern. The `.pi/skills/machine-handoff/SKILL.md` skill generates these on demand.
