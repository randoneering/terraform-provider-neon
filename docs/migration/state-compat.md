# State Compatibility Verification Protocol

Every resource migrating from `terraform-plugin-sdk/v2` to `terraform-plugin-framework`
MUST pass this protocol before its migration release ships. The goal: prove that a
state file written by the SDK v2 version still applies cleanly under the Plugin
Framework version, with no re-import required.

References:

- openspec/changes/migrate-sdk-v2-to-framework/design.md (D7)
- openspec/changes/migrate-sdk-v2-to-framework/specs/plugin-framework-adoption/spec.md
  (Requirement: State compatibility end-to-end)

## Steps

For each migrated resource:

1. **Baseline**: Using the SDK v2 build of the provider (commit on `main` before the
   migration PR), run an acceptance test that creates the resource and produces a
   real Neon-side object. Capture the resulting Terraform state JSON.

2. **Migrate**: Check out the migration PR (Plugin Framework implementation in
   `internal/provider/`). Build the provider binary. Confirm the resource is
   registered under the same Terraform resource type name as the SDK v2 version.

3. **Apply against existing state**: From a directory holding the captured state
   JSON from step 1, run `terraform plan`. Expected output: **no changes**. Any
   drift (attribute renames, type changes, computed vs. required flips) MUST be
   investigated and either fixed in the framework implementation or escalated to
   a spec change.

4. **Apply changes**: Run `terraform apply` (no config edits). The provider MUST
   accept the existing state without prompting for re-import. Plan output must
   remain zero-drift.

5. **Destroy**: Run `terraform destroy`. The Neon-side object must be removed
   identically to the SDK v2 baseline. The state file must transition to empty
   for that resource without orphan records.

## Pass criteria

A migration passes when ALL of the following hold:

- [ ] Step 3 plan output is `No changes. Your infrastructure matches the configuration.`
- [ ] Step 4 apply output is `Apply complete! Resources: 0 added, 0 changed, 0 destroyed.`
- [ ] Step 5 destroy removes the underlying Neon object.
- [ ] No new attributes appear in the post-migration state that weren't present in the
      SDK v2 baseline (with the exception of attributes explicitly added by a
      separate, declared feature gap-fill change).

## Fail handling

If the protocol fails:

1. Open an issue on the migration PR describing the drift.
2. Determine whether the drift is fixable in framework code (re-add the missing
   attribute, fix a type mismatch, etc.) or whether the SDK v2 schema needs
   documenting as the source of truth.
3. Re-run the protocol from step 3.

## Exceptions

The protocol may be skipped ONLY when:

- The resource is brand-new (no SDK v2 baseline exists). In that case the protocol
  starts at step 1 with the framework implementation; the captured state becomes
  the baseline.
- The migration is purely additive (a new attribute on an existing resource, no
  schema changes). In that case steps 3-5 are reduced to a single `terraform
  plan` against a state file that includes the new attribute; plan output must
  still be zero-drift.

## Recording results

Each PR that migrates a resource includes a checklist in the description,
copy-pasted from this doc, with each step dated and the expected outcome recorded
(command output attached or linked from CI artifacts).