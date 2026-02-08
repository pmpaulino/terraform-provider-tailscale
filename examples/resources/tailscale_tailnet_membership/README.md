# tailscale_tailnet_membership example (local provider)

This example uses the **local** provider build so you can run it against your Tailnet without publishing the provider.

**Creating user invites requires a user-owned API key** (e.g. from [Settings > Keys](https://login.tailscale.com/admin/settings/keys)), not OAuth client credentials. See [Tailscale: Invite users](https://tailscale.com/kb/1371/invite-users).

## 1. Build the provider

From the repo root:

```bash
make build
```

This produces `terraform-provider-tailscale` in the repo root.

## 2. Point Terraform at the local provider

Copy `terraformrc.example` to `~/.terraformrc` (or `%APPDATA%\terraform.rc` on Windows). Edit the path so it points at your repo root (the directory that contains the built binary).

```bash
cp terraformrc.example ~/.terraformrc
```

## 3. Set credentials and tailnet

Set environment variables:

- `TAILSCALE_API_KEY` — user-owned API key (with UserInvites and users scope)
- `TAILSCALE_TAILNET` — your tailnet (e.g. `example.com`)
- Optional: `TAILSCALE_BASE_URL` (default `https://api.tailscale.com`)

## 4. Run Terraform

```bash
cd examples/resources/tailscale_tailnet_membership

export TAILSCALE_TAILNET=your-tailnet.com
export TAILSCALE_API_KEY=tskey-api-...

terraform init
terraform plan
# terraform apply   # when ready
```

## Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `base_url` | Tailscale API base URL | `https://api.tailscale.com` |
| `member_login_name` | Email for the member resource | `alice@example.com` |

Replace the default email with a real address in your tailnet, or pass `-var="member_login_name=..."`.

## Import

To import an existing membership:

```bash
terraform import 'tailscale_tailnet_membership.member' 'YOUR_TAILNET:email@example.com'
```
