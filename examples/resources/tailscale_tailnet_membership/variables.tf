# Set TAILSCALE_API_KEY in the environment. Optionally set TAILSCALE_TAILNET and TAILSCALE_BASE_URL,
# or pass tailnet via -var / .tfvars. Creating invites requires a user-owned API key (not OAuth client keys):
# https://tailscale.com/kb/1371/invite-users

variable "base_url" {
  type        = string
  description = "Tailscale API base URL"
  default     = "https://api.tailscale.com"
}

variable "member_login_name" {
  type        = string
  description = "Email for the member membership (invite or existing user)"
  default     = "alice@example.com"
}
