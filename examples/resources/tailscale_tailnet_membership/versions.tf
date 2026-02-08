# Use the local provider build (no version constraint when using dev_overrides).
# See README.md for how to point Terraform at your local provider.
terraform {
  required_providers {
    tailscale = {
      source = "tailscale/tailscale"
    }
  }
  required_version = ">= 1.0"
}
