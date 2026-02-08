# Quickstart: tailscale_tailnet_membership

**Feature**: 001-tailscale-user-management  
**Spec**: [spec.md](./spec.md)

## Prerequisites

- Terraform Tailscale provider configured with credentials (API key or OAuth) that have **UserInvites** and **users** scope.
- Creating invites may require a user-owned API key (see [Tailscale API docs](https://tailscale.com/kb/1371/invite-users)).

## Ensure a member (invite when needed)

Add a membership for an identity. If they are not in the tailnet and have no pending invite, an invitation is sent (e.g. by email). If they are already a member or have a pending invite, the operation is a no-op.

```hcl
resource "tailscale_tailnet_membership" "alice" {
  login_name = "alice@example.com"
  role       = "member"
}
```

With admin role:

```hcl
resource "tailscale_tailnet_membership" "bob_admin" {
  login_name = "bob@example.com"
  role       = "admin"
}
```

## List and inspect

Use the existing data source to list all users (and optionally filter by role). Single membership state is visible on the resource after apply.

```hcl
data "tailscale_users" "all" {}

# After apply, tailscale_tailnet_membership.*.state is pending | active | disabled
output "alice_state" {
  value = tailscale_tailnet_membership.alice.state
}
```

## Disable and re-enable (suspend / restore)

Disable (suspend) is done by updating the resource to a “disabled” state (e.g. via a lifecycle or separate resource that calls suspend). Re-enable by updating back. (Exact attribute name for “disabled” TBD in implementation; e.g. `suspended = true` or state machine.)

```hcl
# Example: disable by setting suspended = true (implementation may differ)
resource "tailscale_tailnet_membership" "alice" {
  login_name = "alice@example.com"
  role       = "member"
  suspended  = true   # or state = "disabled"
}
```

## Remove membership (destroy)

Destroying the resource removes the membership: if the invite is still pending, the invite is cancelled; if the user is already a member, they are removed from the tailnet.

```hcl
# On terraform destroy (or removing the resource), invite is cancelled or user is deleted
```

Optional: downgrade instead of remove (like GitHub’s `downgrade_on_destroy`):

```hcl
resource "tailscale_tailnet_membership" "alice" {
  login_name             = "alice@example.com"
  role                   = "admin"
  downgrade_on_destroy   = true  # On destroy, downgrade to member (or suspend) instead of delete
}
```

## Import

Import an existing membership by tailnet and login_name (email):

```bash
terraform import 'tailscale_tailnet_membership.alice' 'tailnet_xxxxx:alice@example.com'
```

(Exact import ID format TBD in implementation; will match resource ID.)

## Error handling

If the API returns an error (e.g. missing scope, last admin, or rate limit), the provider will surface a clear, actionable message. Ensure your token has **UserInvites** and **users** scope for full membership management.
