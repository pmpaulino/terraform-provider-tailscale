# Implementation Plan: Tailscale User Management (Membership)

**Branch**: `001-tailscale-user-management` | **Date**: 2025-02-07 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `specs/001-tailscale-user-management/spec.md`. Extend the Terraform Tailscale provider with user management following the same flow as the GitHub provider membership resource.

## Summary

Deliver a single Terraform resource that represents **tailnet membership** (one membership per identity). Creating the resource ensures the identity is in the tailnet: if not present, the provider creates a user invite (Tailscale API sends the invitation); if already a member or pending invite, the operation is idempotent. The same resource supports state *pending*, *active*, and *disabled* (Tailscale: suspended). Destroy cancels a pending invite or removes the user (with optional downgrade-on-destroy). Implementation uses the existing Tailscale API (user invites, users list/get, suspend, restore, delete, role update) and the existing `tailscale.com/client/tailscale/v2` client, adding a new resource `tailscale_tailnet_membership` and reusing existing data sources where appropriate.

## Technical Context

**Language/Version**: Go 1.25.x (per go.mod)  
**Primary Dependencies**: hashicorp/terraform-plugin-sdk/v2, tailscale.com/client/tailscale/v2 (v2.7.0)  
**Storage**: N/A (Tailscale Control API is source of truth)  
**Testing**: Go testing + terraform-plugin-sdk; 100% coverage required (constitution)  
**Target Platform**: Terraform 1.x; provider runs in Terraform CLI environment  
**Project Type**: Terraform provider (single module)  
**Performance Goals**: Normal provider apply/read latency; no special targets beyond API responsiveness  
**Constraints**: OAuth scopes must include `UserInvites` and `users` for full membership management; user-owned API keys required for creating invites (per Tailscale API notes)  
**Scale/Scope**: One resource type; tens to hundreds of memberships per tailnet typical

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|--------|
| I. Build Small, Test-Driven Steps | Pass | Implementation will add one resource with Create/Read/Update/Delete; tests first per existing provider pattern. |
| II. Keep Code Simple and Clear | Pass | Single resource file; delegate to Tailscale client; no extra abstraction layers. |
| III. Embrace Feedback and Iteration | Pass | Provider already has feedback via Terraform plan/apply and docs. |
| IV. Automate Relentlessly | Pass | CI already runs tests; no new manual steps. |
| V. Design for Change | Pass | Resource schema and client calls are modular; API changes localized. |
| VI. Optimize for Communication and Learning | Pass | Inline comments and docs for membership ↔ API mapping. |
| VII. Minimal Dependencies | Pass | No new dependencies; use existing tailscale.com client. |
| VIII. Complete Test Coverage | Pass | 100% coverage required; new code in resource + tests only. |

No violations. Complexity tracking table left empty.

## Project Structure

### Documentation (this feature)

```text
specs/001-tailscale-user-management/
├── plan.md              # This file
├── research.md          # Phase 0 (Tailscale API ↔ spec mapping; client capabilities)
├── data-model.md        # Phase 1 (Membership entity, states, Terraform schema)
├── quickstart.md        # Phase 1 (Usage examples)
├── contracts/           # Phase 1 (API operations used)
└── tasks.md             # Phase 2 (/speckit.tasks - not created by plan)
```

### Source Code (repository root)

```text
tailscale/
├── resource_tailnet_membership.go      # NEW: membership resource (Create/Read/Update/Delete)
├── resource_tailnet_membership_test.go # NEW: tests (100% coverage)
├── provider.go                         # UPDATE: register tailscale_tailnet_membership
├── data_source_user.go                # existing
├── data_source_users.go               # existing
├── resource_tailnet_key.go            # existing (reference pattern)
└── ... (other existing resources/data sources)

docs/
├── resources/
│   └── tailnet_membership.md          # NEW: provider docs for tailscale_tailnet_membership
```

**Structure Decision**: The provider is a single Go module under `tailscale/`. New code is one resource file and one test file; provider registration adds one entry. Documentation follows existing `docs/resources/` pattern. No new packages or services.

## Complexity Tracking

*No constitution violations. Table not used.*
