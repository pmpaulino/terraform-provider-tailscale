# Tasks: Tailscale User Management (Membership)

**Input**: Design documents from `specs/001-tailscale-user-management/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md  

**Tests**: Included (constitution requires 100% test coverage and test-first).  

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently. The single resource `tailscale_tailnet_membership` implements all stories; phases break down by CRUD surface (Create/Read → Update → Delete) and docs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- Provider code: `tailscale/` at repository root (absolute: `/Users/pat/Projects/pmpaulino/terraform-provider-tailscale/tailscale/`)
- Docs: `docs/resources/` at repository root
- Specs: `specs/001-tailscale-user-management/`

---

## Phase 1: Setup (Verification)

**Purpose**: Verify Tailscale client and API surface before implementation

- [x] T001 Verify tailscale.com/client/tailscale/v2 exposes or can call UserInvites (list, create, get, delete) and Users (list, get, suspend, restore, delete, update role); document findings or required client extensions in specs/001-tailscale-user-management/research.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Resource schema and membership resolution that all user stories depend on

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete

- [x] T002 Define tailscale_tailnet_membership resource Schema (login_name, role, downgrade_on_destroy, state, invite_id, user_id) and resource ID format (tailnet:login_name) in tailscale/resource_tailnet_membership.go
- [x] T003 Implement membership resolve helper (list user invites + list users for tailnet, find by login_name/email; return invite or user or neither) in tailscale/resource_tailnet_membership.go

**Checkpoint**: Schema and resolve helper ready; Create/Read/Update/Delete can be implemented

---

## Phase 3: User Story 1 - Add or ensure membership (invite when needed) (Priority: P1) 🎯 MVP

**Goal**: Ensure an identity is in the tailnet; create invite when not present; idempotent when already member or pending invite. Read returns state (pending/active) and role.

**Independent Test**: Apply resource with login_name; verify invite is created or state unchanged; run terraform plan again and see no changes; read state/role.

### Tests for User Story 1

- [x] T004 [P] [US1] Add test for Create (ensure membership creates invite when identity not in tailnet) in tailscale/resource_tailnet_membership_test.go
- [x] T005 [P] [US1] Add test for Create idempotency (ensure membership no-op when user or invite already exists) in tailscale/resource_tailnet_membership_test.go
- [x] T006 [P] [US1] Add test for Read (state pending when invite exists; state active when user exists) in tailscale/resource_tailnet_membership_test.go
- [x] T007 [P] [US1] Add test for destroy when pending (cancels invite) in tailscale/resource_tailnet_membership_test.go

### Implementation for User Story 1

- [x] T008 [US1] Implement CreateContext (use resolve helper; if no user and no invite → create user invite with role; else no-op; set ID tailnet:login_name) in tailscale/resource_tailnet_membership.go
- [x] T009 [US1] Implement ReadContext (use resolve helper; set state, role, invite_id or user_id; remove from state if not found for idempotent delete) in tailscale/resource_tailnet_membership.go
- [x] T010 [US1] Wire Resource to Create and Read in tailscale/resource_tailnet_membership.go

**Checkpoint**: User Story 1 complete; ensure membership and read state work; destroy pending cancels invite

---

## Phase 4: User Story 2 - Disable and Re-enable membership (Priority: P2)

**Goal**: Update resource to suspend (disable) or restore (re-enable) user; idempotent when already in desired state.

**Independent Test**: Create membership, accept invite (or use existing user); set suspended/disabled; apply; verify user suspended; set active; apply; verify user restored.

### Tests for User Story 2

- [x] T011 [P] [US2] Add test for Update suspend (disable) and Read state disabled in tailscale/resource_tailnet_membership_test.go
- [x] T012 [P] [US2] Add test for Update restore (re-enable) and Read state active in tailscale/resource_tailnet_membership_test.go
- [x] T013 [P] [US2] Add test for Update idempotent (disable already disabled; re-enable already active) in tailscale/resource_tailnet_membership_test.go

### Implementation for User Story 2

- [x] T014 [US2] Add schema attribute for desired state (e.g. suspended bool or state override) if not already represented by role/state in tailscale/resource_tailnet_membership.go
- [x] T015 [US2] Implement UpdateContext (role change → PATCH user role; transition to disabled → suspend; transition to active → restore; idempotent when unchanged) in tailscale/resource_tailnet_membership.go

**Checkpoint**: User Story 2 complete; disable and re-enable work; idempotent

---

## Phase 5: User Story 3 - Remove membership (Priority: P3)

**Goal**: Destroy resource cancels pending invite or removes user; optional downgrade_on_destroy; idempotent when already absent.

**Independent Test**: Destroy when pending → invite deleted; destroy when active → user deleted; destroy with downgrade_on_destroy → user downgraded or suspended; destroy when already gone → success.

### Tests for User Story 3

- [x] T016 [P] [US3] Add test for Delete when user exists (removes user) in tailscale/resource_tailnet_membership_test.go
- [x] T017 [P] [US3] Add test for Delete idempotent (already removed or invite already deleted) in tailscale/resource_tailnet_membership_test.go
- [x] T018 [P] [US3] Add test for downgrade_on_destroy (on destroy downgrade role or suspend instead of delete) in tailscale/resource_tailnet_membership_test.go

### Implementation for User Story 3

- [x] T019 [US3] Implement DeleteContext (if pending → delete user invite; if user and not downgrade_on_destroy → delete user; if downgrade_on_destroy → PATCH role to member or suspend; handle 404 as success) in tailscale/resource_tailnet_membership.go

**Checkpoint**: User Story 3 complete; remove and optional downgrade work

---

## Phase 6: User Story 4 - List and inspect memberships (Priority: P4)

**Goal**: Read exposes state and role; listing uses existing data source or resource read; docs show how to list and inspect.

**Independent Test**: List users via tailscale_users data source; read single membership resource and verify state/role populated.

### Implementation for User Story 4

- [x] T020 [US4] Ensure ReadContext populates state (pending/active/disabled) and role in tailscale/resource_tailnet_membership.go
- [x] T021 [P] [US4] Add docs for listing (tailscale_users data source) and single membership (resource attributes) in docs/resources/tailnet_membership.md

**Checkpoint**: User Story 4 complete; list and inspect documented

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Provider registration, error handling, last-admin check, docs, coverage

- [x] T022 Register tailscale_tailnet_membership resource in tailscale/provider.go ResourcesMap
- [x] T023 Implement last-admin / account-owner check (FR-009): before disable or delete, ensure not last admin or owner; return clear diag message in tailscale/resource_tailnet_membership.go
- [x] T024 Ensure all API errors surface clear, actionable diag messages (FR-012) in tailscale/resource_tailnet_membership.go
- [x] T025 [P] Add Importer (StateContext) for tailnet:login_name in tailscale/resource_tailnet_membership.go
- [x] T026 [P] Add resource documentation (arguments, attributes, examples, import) in docs/resources/tailnet_membership.md
- [x] T027 Run go test ./tailscale/ -cover and ensure 100% coverage for resource_tailnet_membership.go and resource_tailnet_membership_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1; **blocks** Phases 3–6
- **Phases 3–6 (User Stories)**: Depend on Phase 2; US2 depends on US1 (Update builds on Create/Read); US3 depends on US1 (Delete builds on Create/Read); US4 depends on Read (US1)
- **Phase 7 (Polish)**: Depends on Phases 3–6 complete

### User Story Dependencies

- **US1 (P1)**: After Phase 2; no other story dependency
- **US2 (P2)**: After US1 (UpdateContext needs Read/Create)
- **US3 (P3)**: After US1 (DeleteContext needs Read/Create)
- **US4 (P4)**: After US1 (Read completeness and docs)

### Within Each User Story

- Tests (T004–T007, T011–T013, T016–T018) written and failing before implementation
- Implementation tasks then make tests pass

### Parallel Opportunities

- Phase 1: single task
- Phase 2: T002 and T003 sequential (same file, helper used by schema/resource)
- Phase 3: T004–T007 tests can run in parallel; T008–T010 sequential
- Phase 4: T011–T013 tests parallel; T014–T015 sequential
- Phase 5: T016–T018 tests parallel; T019 sequential
- Phase 6: T020–T021 (T021 [P])
- Phase 7: T025, T026 [P]; T022, T023, T024, T027 sequential or as needed

---

## Parallel Example: User Story 1

```bash
# Write all US1 tests first (parallel):
# T004 Create test, T005 idempotency test, T006 Read test, T007 destroy-pending test
# Then implement T008 Create, T009 Read, T010 wire Resource
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Verify client API
2. Phase 2: Schema + resolve helper
3. Phase 3: US1 tests then Create/Read
4. **STOP and VALIDATE**: terraform apply with new resource; verify invite or no-op; terraform plan no changes
5. Add US2–US4 and Polish as needed

### Incremental Delivery

1. Phase 1 + 2 → foundation
2. Phase 3 (US1) → MVP (ensure + read + destroy pending)
3. Phase 4 (US2) → disable/re-enable
4. Phase 5 (US3) → full delete + downgrade_on_destroy
5. Phase 6 (US4) → docs for list/inspect
6. Phase 7 → register, errors, docs, coverage

### Suggested MVP Scope

- **MVP**: Phases 1–3 (Setup, Foundational, User Story 1). Delivers: create membership (invite when needed), read state/role, destroy cancels invite. Enough to add/invite users and remove pending invites.

---

## Notes

- [P] tasks = different files or independent test cases
- [Story] label maps task to spec user story for traceability
- Single resource file implements all stories; phases split by CRUD and docs
- Commit after each task or logical group
- Constitution: 100% test coverage; tests first
