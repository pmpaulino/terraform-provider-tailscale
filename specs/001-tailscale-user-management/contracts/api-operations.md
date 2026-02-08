# API Operations Used: Tailscale Control API

**Feature**: 001-tailscale-user-management  
**Source**: tailscale-api.json (Tailscale Control API)

This document lists the API operations the `tailscale_tailnet_membership` resource depends on. The provider uses `tailscale.com/client/tailscale/v2`; these operations must be available on that client (or implemented via the same HTTP client).

## User Invites (scope: UserInvites)

| Operation | Method | Path | Purpose |
|-----------|--------|------|---------|
| List user invites | GET | `/tailnet/{tailnet}/user-invites` | Find pending invite by email for Read/Create idempotency |
| Create user invite | POST | `/tailnet/{tailnet}/user-invites` | Ensure membership for new identity (invite with email, role) |
| Get user invite | GET | `/user-invites/{userInviteId}` | Optional: read single invite |
| Delete user invite | DELETE | `/user-invites/{userInviteId}` | Cancel invite on resource destroy when state is pending |

Create body: `{ "email"?: string, "role"?: "member" | "admin" | ... }` (role default member). Response includes invite ID and inviteUrl.

## Users (scope: users)

| Operation | Method | Path | Purpose |
|-----------|--------|------|---------|
| List users | GET | `/tailnet/{tailnet}/users` | Find user by login_name; list members for Read |
| Get user | GET | `/users/{userId}` | Read single user (existing data source) |
| Update user role | PATCH | `/users/{userId}/role` | Update role (member/admin); optional downgrade on destroy |
| Suspend user | POST | `/users/{userId}/suspend` | Disable membership (FR-002) |
| Restore user | POST | `/users/{userId}/restore` | Re-enable membership (FR-003) |
| Delete user | POST | `/users/{userId}/delete` | Remove membership (FR-004) |

Note: Suspend, restore, delete are documented as Personal/Enterprise. Provider surfaces API errors with clear messages (FR-012).

## Resource → API flow (summary)

- **Create**: List invites + list users; if login_name in users → no-op; if in invites → no-op; else POST create user invite.
- **Read**: List invites + list users; find by login_name; set state (pending | active | disabled from user status).
- **Update**: If role changed and user exists → PATCH role. If state transition disabled ↔ active → suspend or restore.
- **Delete**: If pending → DELETE invite. If user and not downgrade_on_destroy → POST delete user. If downgrade_on_destroy → PATCH role to member or POST suspend (chosen semantics).
