# Handoff — framework-migration-pilot

**Last verified:** 2026-08-31
**Repo:** randoneering/terraform-provider-neon (fork of neondatabase:main)

## Branch state

Local:
- `feat/framework-migration-pilot` — clean, in sync with origin, PR open against upstream
- `openspec/framework-migration-pilot` — clean, in sync with origin, openspec artifacts only

Remote:
- `origin/feat/framework-migration-pilot` — PR open
- `origin/openspec/framework-migration-pilot` — openspec artifacts backup
- `origin/chore/pr-template` — separate PR for the PR template

## OpenSpec change status

Active change: `framework-migration-pilot` (in-progress, 11/14 tasks complete)

Done: 1.1, 1.2, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 3.4, 5.2, 6.1
Pending: 4.1, 4.2, 5.1

## Done this session

Five SSH-signed Conventional Commits pushed to `feat/framework-migration-pilot`:
- `f8951c5 docs(changelog): note framework pilot pre-release`
- `0e77d56 docs(migration): add user migration guide`
- `7521eff test(provider): add pilot acceptance tests`
- `cde8402 chore(provider): add framework binary entry`
- `9b16821 feat(provider): pilot neon_api_key on framework`

Plus four rebased pilot commits and one openspec commit on the openspec branch.

FSM acceptance tests verified against real Neon — all 4 cases PASS in 15.87s.

## Environment expectations on the next machine

- `NEON_API_KEY` set (Personal API Token from console.neon.tech)
- `TF_ACC=1` to enable acceptance tests
- SSH agent has `randoneeringkey` loaded for SSH commit signing
- Go toolchain supports terraform-plugin-framework v1.19+

## Resume from here

1. Fetch and checkout:

   ```bash
   git clone git@github.com:randoneering/terraform-provider-neon.git
   cd terraform-provider-neon
   git fetch --all --prune
   git checkout feat/framework-migration-pilot
   git checkout openspec/framework-migration-pilot  # for openspec artifacts
   ```

2. Verify state:

   ```bash
   openspec list --json
   openspec validate framework-migration-pilot --strict
   go build ./... && go vet ./...
   ```

3. Re-run FSM tests against real Neon:

   ```bash
   TF_ACC=1 NEON_API_KEY=$NEON_API_KEY go test -count=1 -run TestAccAPIKeyFrameworkFSM -v ./internal/provider/
   ```

4. Continue with pending tasks:
   - **4.1** schema-diff unit test (recommended, no creds needed)
   - **4.2** capture SDK v2 state file, then plan against framework
   - **5.1** cut `v0.16.0-pre.1` tag locally, push when creds available

## Files of interest

Code:
- `internal/provider/resource_api_key.go` — Framework port with FSM retry wiring
- `internal/provider/retry.go` — FSM retry helper mirroring `provider/retry.go`
- `internal/provider/provider.go` — Framework provider with one resource
- `internal/provider/resource_api_key_acc_test.go` — FSM acceptance tests
- `cmd/neon-framework/main.go` — Binary entry serving `registry.terraform.io/neon/neon`
- `CHANGELOG.md` — `v0.16.0-pre.1` entry

Docs:
- `docs/migration/kislerdm-to-neon.md` — user-facing migration guide

Planning (on the `openspec/framework-migration-pilot` branch):
- `openspec/changes/framework-migration-pilot/proposal.md`
- `openspec/changes/framework-migration-pilot/design.md`
- `openspec/changes/framework-migration-pilot/tasks.md`
- `openspec/changes/framework-migration-pilot/specs/framework-hosted-provider/spec.md`

## Known caveats

- `openspec/` is gitignored at the repo root. Workaround on dedicated branches: `git add -f openspec/`.
- SSH signing produces signatures but local verification needs `gpg.ssh.allowedSignersFile` configured. GitHub can verify without this.
- Framework port has no `ImportState` method — `terraform import` fails with a clear error, consistent with SDK v2 behavior.
- Two-namespace state (`kislerdm/neon` and `neon/neon`) is transitional; destination is single `neon/neon` after Phase 0 governance + Phase 3 bulk migration.

## Reusable patterns established this session

- **Openspec branch on fork.** Force-add `openspec/` artifacts to a dedicated `openspec/*` branch so they live on the fork but stay out of feature-branch PRs (which are openspec-free because `.gitignore` excludes the directory).
- **SSH signing.** `commit.gpgsign=true`, `gpg.format=ssh`, `user.signingkey=~/.ssh/randoneeringkey.pub` — already configured in this environment.
- **Conventional Commits scoped.** Use `feat(scope):`, `chore(scope):`, `docs(scope):` with subject lines of a few words.
- **Two-namespace model.** `cmd/neon-framework/` binary serves `registry.terraform.io/neon/neon`; SDK v2 binary serves `registry.terraform.io/kislerdm/neon`. Documented in `design.md` Migration Handoff section.
