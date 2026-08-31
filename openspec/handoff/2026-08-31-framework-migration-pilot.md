# Handoff — framework-migration-pilot

**Last verified:** 2026-08-31
**Repo:** randoneering/terraform-provider-neon (fork of neondatabase:main)

## Branch state

Local + remote:
- `main` — at `b34baa1`, includes the merge of PR #3
- `feat/framework-migration-pilot-4.1` — merged via PR #3, safe to delete
- Tag `v0.16.0-pre.1` (`0879260`) pushed to origin

## OpenSpec change status

Change `framework-migration-pilot` is **archived**:
- 14/14 tasks complete
- Delta spec synced to `openspec/specs/framework-hosted-provider/spec.md`
- Moved to `openspec/changes/archive/2026-08-31-framework-migration-pilot/`
- `openspec validate --all --strict` passes
- `openspec list` shows zero active changes

## Done this session

Six SSH-signed Conventional Commits landed on `main`:

```
b34baa1 chore(openspec): mark 5.1 done after cutting v0.16.0-pre.1
a3f5680 Merge pull request #3 from randoneering/feat/framework-migration-pilot-4.1
5b7d561 docs(openspec): narrow 4.1 wording to schema-compat scope
f5b6339 test(provider): add api_key state-compat integration test
b21eec3 test(provider): add api_key schema-compat test
```

Plus tag `v0.16.0-pre.1` (annotated, SSH-signed).

**Verification**
- `go test ./...` — clean
- Schema-compat test (`provider/resource_api_key_schema_test.go`) passes against both SDK v2 and framework schemas; verified to catch drift
- State-compat integration test (`tests/state-compat/state_compat_test.go`) passes against real Neon: SDK v2 create → framework plan reports "No changes" with exit 0
- CodeRabbit review on PR #3 resolved (one wording tweak on tasks.md)

## Environment expectations on the next machine

- `NEON_API_KEY` set (Personal API Token from console.neon.tech)
- `TF_ACC=1` to enable acceptance and state-compat tests
- SSH agent has `randoneeringkey` loaded for SSH commit signing
- `TERRAFORM_BIN` if terraform is not on PATH (default lookup uses `terraform`)
- Go toolchain supports terraform-plugin-framework v1.19+
- terraform CLI v1.13+ for `-detailed-exitcode` flag

## Resume from here

```bash
git clone git@github.com:randoneering/terraform-provider-neon.git
cd terraform-provider-neon
git checkout main
git fetch --all --prune --tags

openspec list                       # expect 0 active changes
openspec validate --all --strict    # expect pass

go build ./...
go test -count=1 ./...

# With creds:
TF_ACC=1 NEON_API_KEY=$NEON_API_KEY go test -count=1 ./...
TF_ACC=1 NEON_API_KEY=$NEON_API_KEY go test -count=1 -v -run TestStateCompat_SDKv2ToFramework ./tests/state-compat/...
```

## Files of interest

Code:
- `internal/provider/resource_api_key.go` — Framework port with FSM retry wiring
- `internal/provider/retry.go` — FSM retry helper mirroring `provider/retry.go`
- `internal/provider/provider.go` — Framework provider with one resource
- `internal/provider/resource_api_key_acc_test.go` — FSM acceptance tests (refresh/destroy/update/import)
- `provider/resource_api_key_schema_test.go` — schema-compat unit test
- `tests/state-compat/state_compat_test.go` — state-compat integration test (drives terraform CLI via dev_overrides)
- `cmd/neon-framework/main.go` — Binary entry serving `registry.terraform.io/neon/neon`

Tests use:
- `t.TempDir()` for isolation
- Single `binDir` shared across phases; binary is copied in/out for SDK v2 ↔ framework swap (avoids re-init registry query)
- `t.Cleanup` destroys via SDK v2 to guarantee API key revocation

Docs:
- `docs/migration/kislerdm-to-neon.md` — user-facing migration guide
- `CHANGELOG.md` — `v0.16.0-pre.1` entry
- `.goreleaser.yml` — release pipeline config (no `.github/workflows/release.yml` exists; tag push does not auto-build)

Planning (archived):
- `openspec/changes/archive/2026-08-31-framework-migration-pilot/{proposal,design,tasks}.md`
- `openspec/changes/archive/2026-08-31-framework-migration-pilot/specs/framework-hosted-provider/spec.md`

Main specs:
- `openspec/specs/framework-hosted-provider/spec.md` — 4 requirements synced from delta

## Known caveats

- **No release workflow.** Pushing a tag does not trigger goreleaser; the `.goreleaser.yml` is configured but `.github/workflows/release.yml` is missing. Tag `v0.16.0-pre.1` is on origin but no artifact was published automatically.
- **Framework port has no `ImportState`** — `terraform import` fails with a clear error, consistent with SDK v2 behavior. Acceptable per the archived spec.
- **Two-namespace state (`kislerdm/neon` ↔ `neon/neon`)** is transitional; archived spec covers it (Scenario: "State plans unchanged across namespace handoff"). The state-compat test exercises the address-level handoff.
- **State-compat test cannot re-init** — registry query for `neon/neon` fails because no published versions exist. Workaround: swap the binary in the same `dev_overrides` dir without re-init.
- **SSH commit signature verification** needs `gpg.ssh.allowedSignersFile` configured (this repo sets it locally to `~/.ssh/allowed_signers`). GitHub verifies without it.

## Reusable patterns established this session

- **Single-namespace provider addresses.** Both SDK v2 and framework binaries serve `registry.terraform.io/neon/neon`. The handoff's "two-namespace" note reflects the historical registry state, not this fork's binaries.
- **dev_overrides binary swap.** Same `binDir`, copy the target binary in before each phase. Avoids re-init's registry query.
- **Conventional Commits scoped.** `feat(scope):`, `chore(scope):`, `docs(scope):`, `test(scope):` with subject lines of a few words.
- **Feature branch + signed PR + merge back.** Working branches stay short-lived; main is the integration point.
- **openspec/ is tracked in this repo** (unlike the handoff's earlier note). No force-add needed for spec/handoff files.
- **TDD-style schema tests.** The unit test caught drift when `name.Required` was flipped to `false` (manually verified). Cheap insurance against accidental schema mutations during future migrations.
