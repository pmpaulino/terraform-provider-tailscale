# Tailscale Tailnet Membership example (local provider).
# Credentials: TAILSCALE_API_KEY (user-owned key required for creating invites), TAILSCALE_TAILNET.

resource "tailscale_tailnet_membership" "member" {
  login_name = var.member_login_name
  role       = "member"
}

data "tailscale_users" "all" {}

output "member_state" {
  value       = tailscale_tailnet_membership.member.state
  description = "State of the member: pending, active, or disabled"
}

output "users_count" {
  value       = length(data.tailscale_users.all.users)
  description = "Number of users in the tailnet"
}
