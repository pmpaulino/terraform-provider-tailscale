# Credentials from environment: TAILSCALE_API_KEY, TAILSCALE_TAILNET (optional: TAILSCALE_BASE_URL).
provider "tailscale" {
  base_url = var.base_url
}
