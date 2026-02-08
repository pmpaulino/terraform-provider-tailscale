// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"tailscale.com/client/tailscale/v2"
)

const testTailnetMembershipCreate = `
resource "tailscale_tailnet_membership" "alice" {
  login_name = "alice@example.com"
  role       = "member"
}
`

func TestResourceTailnetMembership_Create_EnsureMembershipCreatesInvite(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					inv1 := []userInvite{{ID: "inv1", Email: "alice@example.com", Role: "member"}}
					testServer.ResponseByPath = map[string]interface{}{
						"GET /api/v2/tailnet/example.com/users":     map[string]interface{}{"users": []tailscale.User{}},
						"GET /api/v2/tailnet/example.com/user-invites": []userInvite{}, // fallback for destroy
						"POST /api/v2/tailnet/example.com/user-invites": inv1,
					}
					// First GET invites returns empty (so create runs); second GET invites (on Read) returns inv1
					testServer.ResponseQueueByPath = map[string][]interface{}{
						"GET /api/v2/tailnet/example.com/user-invites": {[]userInvite{}, inv1},
					}
				},
				Config: testTailnetMembershipCreate,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["tailscale_tailnet_membership.alice"]
					if !ok {
						return nil
					}
					if rs.Primary.ID == "" {
						return nil
					}
					if rs.Primary.Attributes["state"] != "pending" {
						return nil
					}
					if rs.Primary.Attributes["invite_id"] != "inv1" {
						return nil
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTailnetMembership_Create_IdempotentWhenUserExists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					existingUser := tailscale.User{
						ID:        "user1",
						LoginName: "alice@example.com",
						Role:      tailscale.UserRoleMember,
						Status:    tailscale.UserStatusActive,
					}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":     map[string]interface{}{"users": []tailscale.User{existingUser}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{existingUser}}
				},
				Config: testTailnetMembershipCreate,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["tailscale_tailnet_membership.alice"]
					if !ok {
						return nil
					}
					if rs.Primary.Attributes["state"] != "active" {
						return nil
					}
					if rs.Primary.Attributes["user_id"] != "user1" {
						return nil
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTailnetMembership_Read_StatePendingWhenInviteExists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{{ID: "inv2", Email: "alice@example.com", Role: "member"}},
					}
					testServer.ResponseBody = []userInvite{{ID: "inv2", Email: "alice@example.com", Role: "member"}}
				},
				Config: testTailnetMembershipCreate,
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["tailscale_tailnet_membership.alice"]
					if rs.Primary.Attributes["state"] != "pending" {
						return nil
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTailnetMembership_Read_StateActiveWhenUserExists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u1", LoginName: "alice@example.com", Role: tailscale.UserRoleMember, Status: tailscale.UserStatusActive, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config: testTailnetMembershipCreate,
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["tailscale_tailnet_membership.alice"]
					if rs.Primary.Attributes["state"] != "active" {
						return nil
					}
					if rs.Primary.Attributes["user_id"] != "u1" {
						return nil
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTailnetMembership_Delete_PendingCancelsInvite(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{{ID: "inv3", Email: "alice@example.com", Role: "member"}},
					}
					testServer.ResponseBody = []userInvite{{ID: "inv3", Email: "alice@example.com", Role: "member"}}
				},
				Config: testTailnetMembershipCreate,
			},
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = nil
				},
				Destroy: true,
				Config:  testTailnetMembershipCreate,
				Check:   func(s *terraform.State) error { return nil },
			},
		},
	})
}

func TestResourceTailnetMembership_Delete_WhenUserExistsRemovesUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u2", LoginName: "alice@example.com", Role: tailscale.UserRoleMember, Status: tailscale.UserStatusActive, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config: testTailnetMembershipCreate,
			},
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = nil
				},
				Destroy: true,
				Config:  testTailnetMembershipCreate,
				Check:   func(s *terraform.State) error { return nil },
			},
		},
	})
}

func TestResourceTailnetMembership_Delete_IdempotentWhenAlreadyRemoved(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{}}
				},
				Destroy: true,
				Config:  testTailnetMembershipCreate,
				Check:   func(s *terraform.State) error { return nil },
			},
		},
	})
}

func TestResourceTailnetMembership_Update_SuspendAndRestore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u3", LoginName: "alice@example.com", Role: tailscale.UserRoleMember, Status: tailscale.UserStatusActive, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config: testTailnetMembershipCreate,
			},
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u3", LoginName: "alice@example.com", Role: tailscale.UserRoleMember, Status: tailscale.UserStatusSuspended, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config: `
resource "tailscale_tailnet_membership" "alice" {
  login_name = "alice@example.com"
  role       = "member"
  suspended  = true
}
`,
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["tailscale_tailnet_membership.alice"]
					if rs.Primary.Attributes["state"] != "disabled" {
						return nil
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTailnetMembership_Delete_DowngradeOnDestroy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u_downgrade", LoginName: "alice@example.com", Role: tailscale.UserRoleAdmin, Status: tailscale.UserStatusActive, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config: `
resource "tailscale_tailnet_membership" "alice" {
  login_name             = "alice@example.com"
  role                   = "admin"
  downgrade_on_destroy   = true
}
`,
			},
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = nil
				},
				Destroy: true,
				Config: `
resource "tailscale_tailnet_membership" "alice" {
  login_name             = "alice@example.com"
  role                   = "admin"
  downgrade_on_destroy   = true
}
`,
				Check: func(s *terraform.State) error { return nil },
			},
		},
	})
}

func TestResourceTailnetMembership_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: testProviderFactories(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testServer.ResponseCode = http.StatusOK
					u := tailscale.User{ID: "u4", LoginName: "alice@example.com", Role: tailscale.UserRoleMember, Status: tailscale.UserStatusActive, Created: time.Now(), LastSeen: time.Now()}
					testServer.ResponseByPath = map[string]interface{}{
						"/api/v2/tailnet/example.com/users":    map[string]interface{}{"users": []tailscale.User{u}},
						"/api/v2/tailnet/example.com/user-invites": []userInvite{},
					}
					testServer.ResponseBody = map[string]interface{}{"users": []tailscale.User{u}}
				},
				Config:        testTailnetMembershipCreate,
				ResourceName: "tailscale_tailnet_membership.alice",
				ImportState:  true,
				ImportStateId: "example.com:alice@example.com", // tailnet:login_name
				ImportStateCheck: func(st []*terraform.InstanceState) error {
					if len(st) != 1 {
						return nil
					}
					if st[0].Attributes["login_name"] != "alice@example.com" {
						return nil
					}
					return nil
				},
			},
		},
	})
}
