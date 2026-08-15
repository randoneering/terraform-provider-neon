# Pilot Migration Checklist

Reusable checklist for migrating a single resource from `terraform-plugin-sdk/v2`
to `terraform-plugin-framework`. The pilot was `neon_api_key` (see
`internal/provider/resource_api_key.go`); this document captures the pattern so
the bulk phase (tasks 3.1-3.10) can repeat it without re-deriving the shape.

## Pre-flight

- [ ] Read the SDK v2 implementation in `provider/resource_<name>.go`
- [ ] Read the existing acceptance tests in `provider/acc-<name>_test.go` and
      `provider/resource_<name>_test.go` (where they exist)
- [ ] Read `openspec/changes/migrate-sdk-v2-to-framework/design.md` (D1-D7)
- [ ] Confirm the resource is registered in `provider/provider.go` ResourcesMap

## Scaffold

- [ ] Add `internal/provider/resource_<name>.go` with the framework resource
      struct that embeds `*neonProvider` (so it can reach `provider.client`)
- [ ] Define the resource model (a Go struct with `tfsdk:"..."` tags matching the
      SDK v2 attribute names and types)
- [ ] Implement `Metadata`, `Schema`, `Create`, `Read`, `Update`, `Delete`
- [ ] For computed attributes that map to existing state, add `UseStateForUnknown`
      or `UseStateForUnknown` plan modifiers where appropriate
- [ ] For force-new semantics on the SDK v2 side, add `RequiresReplace` plan modifier

## Schema parity check

Before wiring the framework resource into the provider:

- [ ] Attribute names match the SDK v2 resource exactly
- [ ] Attribute types match (string, bool, int64, list, map, etc.)
- [ ] Required / Optional / Computed / Sensitive flags match
- [ ] ForceNew flags map to `RequiresReplace` plan modifiers
- [ ] Description text matches the SDK v2 Description (or accepts the new copy if
      the spec explicitly permits it)
- [ ] Nested blocks (TypeList/TypeSet/TypeMap in SDK v2) map to framework
      `NestedBlockObject` or `AttributesNested` correctly

## Wire into provider

- [ ] Add the new framework resource to `internal/provider/provider.go` Resources()
      list **behind the SDK v2 registration in `provider/provider.go`**
- [ ] Note: during the per-resource migration phase, the SDK v2 provider still
      serves the un-migrated resources. The framework `internal/provider/` is a
      parallel package, not a replacement, until all resources migrate
- [ ] The final cutover (rewriting `main.go` to serve `internal/provider/`)
      happens once every resource has been migrated and the SDK v2 `provider/`
      package is empty

## Tests

- [ ] Port the existing acceptance test to `internal/provider/resource_<name>_test.go`
      using `terraform-plugin-testing` (`acctest`, `resource.Test`)
- [ ] Verify the test compiles: `go test ./internal/provider/ -count=1 -run "^$"`
- [ ] Run the protocol in `docs/migration/state-compat.md` end-to-end
- [ ] Confirm the post-migration acceptance test passes against a real Neon project

## Release

- [ ] Open a PR; reference the design.md decision that justified the migration
      ordering (e.g. for `neon_api_key`: D1 pilot)
- [ ] In the PR description, attach the state-compat protocol output
- [ ] After review + merge, tag and ship a release with the migrated resource
- [ ] Confirm no SDK v2 churn for un-migrated resources in the same release

## Anti-patterns to avoid

- **Big-bang moves**: do NOT move the entire `provider/` package to
  `internal/provider/` in one PR. Per-resource is the discipline.
- **Schema drift**: do NOT rename attributes or change types as part of the
  migration; that breaks state compat. Schema changes belong in a separate change.
- **Logic changes**: do NOT change retry behavior, error wrapping, or async
  operation handling as part of the migration. Mechanical translation only.
- **Test coverage reduction**: do NOT drop SDK v2 baseline scenarios from the
  ported test. Same count or more.

## Pilot retrospective (neon_api_key)

After `neon_api_key` ships, fill in:

- Lines of code: SDK v2 `provider/resource_api_key.go` ~120 LoC vs framework
  `internal/provider/resource_api_key.go` ~165 LoC
- Test parity: same scenarios (create, read, delete), no test regression
- Build impact: terraform-plugin-framework v1.19.0 added; SDK v2 retained for
  un-migrated resources
- Time to first green test: TBD on first CI run