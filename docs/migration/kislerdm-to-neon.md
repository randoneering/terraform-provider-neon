# Migrating from `kislerdm/neon` to `neon/neon`

This document describes the user-facing migration path from the legacy
`kislerdm/neon` Terraform provider (built on `terraform-plugin-sdk/v2`) to the
new `neon/neon` provider (built on `terraform-plugin-framework`).

## When to migrate

The two namespaces coexist during the migration window. Migrate your
configuration once **all** of the following are true:

- `neon/neon` ships with the full resource and data source surface you use
  (11 resources, 5 data sources). Check the [release notes](#release-coverage)
  for the latest coverage map.
- The deprecation notice for `kislerdm/neon` has been published in the v0.x
  release line and either a Terraform Registry redirect is in place or you are
  ready to edit `source = "..."` in every Terraform configuration.

If you use only `neon_api_key`, you can adopt `neon/neon` immediately on any
`v0.16.0-pre.*` pre-release.

## What changes for you

One line in your `required_providers` block:

```hcl
# Before
terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.15"
    }
  }
}

# After
terraform {
  required_providers {
    neon = {
      source  = "neon/neon"
      version = ">= 1.0.0"
    }
  }
}
```

Nothing else.

## What does not change

- **State files.** Terraform state files reference resources by their resource
  address (for example, `neon_api_key.example`), not by the provider source.
  Your existing state is byte-for-byte compatible with `neon/neon` as long as
  the resource schema has not changed.
- **Resource addresses.** Every `neon_api_key`, `neon_project`, `neon_branch`,
  and so on keeps the same address.
- **Attribute names and types.** The schema is identical across the two
  namespaces; see [State Compatibility Verification Protocol](state-compat.md).

## Pre-migration check

Before flipping the source line, verify that `neon/neon` covers every resource
and data source you use:

```bash
# From the repo root, with both binaries on PATH:
terraform providers schema -json \
  | jq -r '.provider_schemas | to_entries[] | "\(.key): \(.value.resource_schemas | keys | join(", "))"'
```

Compare the output against what your configuration references. Any resource
listed in your config but missing from `neon/neon` blocks the migration until
that resource's Phase 3 migration lands.

## Migration steps

1. **Update the source line.** Edit `terraform { required_providers { neon.source } }`
   in every Terraform configuration that uses the provider.

2. **Reinitialize.** From each working directory:

   ```bash
   terraform init -upgrade
   ```

   Terraform downloads the new provider binary and updates the lock file. You
   should see output indicating the provider was downloaded from
   `registry.terraform.io/neon/neon`.

3. **Plan with the existing state.** Without changing any other configuration:

   ```bash
   terraform plan
   ```

   Expected output:

   ```
   No changes. Your infrastructure matches the configuration.
   ```

   Anything else (a non-empty plan, attribute drift, replacement notices) means
   the schemas have diverged. Stop and file an issue against the migration
   change; do not proceed to apply.

4. **Apply.** With an empty plan confirmed:

   ```bash
   terraform apply
   ```

   Expected output:

   ```
   Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
   ```

5. **Repeat for every working directory.** Each Terraform configuration that
   references `neon/neon` needs the same three steps.

## Rollback

If something goes wrong after step 4, the rollback path is also a one-line
config change:

```hcl
source = "kislerdm/neon"
```

Then `terraform init -upgrade` and `terraform plan`. The state file is
unaffected because resource addresses did not change. Resources that were
created or modified under `neon/neon` remain visible to `kislerdm/neon` and
vice versa, since the underlying Neon API is the same.

## Release coverage

| Provider version | Resources on `neon/neon` |
|------------------|--------------------------|
| `v0.16.0-pre.1`  | `neon_api_key`           |
| `v0.17.0-pre.*`  | `neon_api_key`, ...      |
| `v1.0.0`         | All 11 resources + 5 data sources |

The full coverage map updates with each release. Check the [CHANGELOG](../../CHANGELOG.md)
for the current list of resources and data sources available under `neon/neon`.

## See also

- [State Compatibility Verification Protocol](state-compat.md) — how state
  compatibility is verified during migration, and the protocol each migrated
  resource passes before its release ships.
- [Pilot Checklist](pilot-checklist.md) — internal checklist for contributors
  migrating a resource.
