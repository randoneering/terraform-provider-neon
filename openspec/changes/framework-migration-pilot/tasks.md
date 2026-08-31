## 1. Setup

- [x] 1.1 Rebase `feat/migrate-sdk-v2-to-framework` onto `c81958f` and verify the rebase resolves with conflicts only inside `provider/resource_api_key.go` and `go.mod`
- [x] 1.2 Verify the rebased branch builds by running `go build ./...` and confirming zero errors

## 2. Pilot Implementation

- [x] 2.1 Port `resource_api_key.go` to Plugin Framework and verify it implements the framework `resource.Resource` interface (`Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`)
- [x] 2.2 Port FSM handlers (`ReadRetry`, `DeleteRetry`, `Import`) into the framework implementation and verify each handler reproduces the SDK v2 behavior on the four cases (refresh, destroy, update, import)
- [x] 2.3 Confirm both SDK v2 and framework resources register in `provider.go` and verify `terraform providers` lists the provider without error

## 3. Acceptance Tests

- [x] 3.1 Add acceptance test for the refresh case and verify the test passes against a real Neon project
- [x] 3.2 Add acceptance test for the destroy case including out-of-band pre-deletion and verify the test passes
- [x] 3.3 Add acceptance test for the update case and verify the test passes
- [x] 3.4 Add acceptance test for the import case and verify the test produces a valid state

## 4. State Compatibility

- [x] 4.1 Add a schema-compat unit test that asserts the framework port of `neon_api_key` exposes the same attribute names, types, and Required/Optional/Computed/Sensitive flags as the SDK v2 implementation
- [x] 4.2 Run the state-compat integration test that applies with the SDK v2 implementation and plans against the framework implementation, and verify the plan output is empty (no diff)

## 5. Release

- [ ] 5.1 Cut a pre-release tag (for example `v0.16.0-pre.1`) and verify the build artifact publishes successfully
- [x] 5.2 Update `CHANGELOG.md` to note the framework pilot is available as a pre-release and verify the entry reads clearly

## 6. User-Facing Migration Documentation

- [x] 6.1 Create `docs/migration/kislerdm-to-neon.md` describing the user-facing migration path from `kislerdm/neon` to `neon/neon` (one-line `source` change, state files unchanged, when to migrate, rollback) and verify the doc reads cleanly with example configs that pass `terraform validate`
