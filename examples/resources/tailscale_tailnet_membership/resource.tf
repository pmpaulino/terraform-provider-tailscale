resource "tailscale_tailnet_membership" "member" {
  login_name = "alice@example.com"
  role       = "member"
}
