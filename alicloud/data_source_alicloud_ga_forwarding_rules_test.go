package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAlicloudGaForwardingRulesDataSource(t *testing.T) {
	rand := acctest.RandInt()
	resourceId := "data.alicloud_ga_forwarding_rules.default"
	testAccConfig := dataSourceTestAccConfigFunc(resourceId, "", dataSourceGaForwardingRulesConfigDependence)
	checkoutSupportedRegions(t, true, connectivity.GaSupportRegions)
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}"},
		}),
		fakeConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}_fake"},
		}),
	}

	statusConf := dataSourceTestAccConfig{
		existConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}"},
			"status":         "active",
		}),
		fakeConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}_fake"},
			"status":         "configuring",
		}),
	}

	allConf := dataSourceTestAccConfig{
		existConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}"},
			"status":         "active",
		}),
		fakeConfig: testAccConfig(map[string]interface{}{
			"accelerator_id": "${local.ga_accelerator_id}",
			"listener_id":    "${alicloud_ga_listener.example.id}",
			"ids":            []string{"${alicloud_ga_forwarding_rule.default.forwarding_rule_id}_fake"},
			"status":         "configuring",
		}),
	}

	var existMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                                     "1",
			"forwarding_rules.#":                        "1",
			"forwarding_rules.0.priority":               "1",
			"forwarding_rules.0.forwarding_rule_id":     CHECKSET,
			"forwarding_rules.0.forwarding_rule_name":   "",
			"forwarding_rules.0.forwarding_rule_status": "active",
			"forwarding_rules.0.listener_id":            CHECKSET,
			"forwarding_rules.0.rule_conditions.#":      "1",
			"forwarding_rules.0.rule_actions.#":         "1",
			"forwarding_rules.0.id":                     CHECKSET,
		}
	}

	var fakeMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"forwarding_rules.#": "0",
			"ids.#":              "0",
		}
	}

	var CheckInfo = dataSourceAttr{
		resourceId:   resourceId,
		existMapFunc: existMapFunc,
		fakeMapFunc:  fakeMapFunc,
	}
	preCheck := func() {}

	CheckInfo.dataSourceTestCheckWithPreCheck(t, rand, preCheck, idsConf, statusConf, allConf)
}

func dataSourceGaForwardingRulesConfigDependence(name string) string {
	return fmt.Sprintf(`
locals {
  ga_accelerator_id = length(data.alicloud_ga_accelerators.default.ids) > 0 ? data.alicloud_ga_accelerators.default.ids[0] : concat(alicloud_ga_accelerator.default.*.id, [""])[0]
  ga_bandwidth_id   = length(data.alicloud_ga_bandwidth_packages.default.ids) > 0 ? data.alicloud_ga_bandwidth_packages.default.ids[0] : concat(alicloud_ga_bandwidth_package.default.*.id, [""])[0]
}
data "alicloud_ga_accelerators" "default" {}
data "alicloud_ga_bandwidth_packages" "default" {}

resource "alicloud_ga_accelerator" "default" {
  count            = length(data.alicloud_ga_accelerators.default.ids) > 0 ? 0 : 1
  duration         = 1
  spec             = "1"
  accelerator_name = "tf-testaccGa_accelerator"
  auto_use_coupon  = false
  description      = "tf-testaccGa_accelerator"
}

resource "alicloud_ga_bandwidth_package" "default" {
  count                  = length(data.alicloud_ga_bandwidth_packages.default.ids) > 0 ? 0 : 1
  bandwidth              = 100
  type                   = "Basic"
  bandwidth_type         = "Basic"
  payment_type           = "PayAsYouGo"
  billing_type           = "PayBy95"
  ratio                  = 30
  bandwidth_package_name = "%s"
}

resource "alicloud_ga_listener" "example" {
  accelerator_id = local.ga_accelerator_id
  port_ranges {
    from_port = 70
    to_port   = 70
  }
  protocol = "HTTP"
}

resource "alicloud_eip_address" "example" {
  bandwidth            = "10"
  internet_charge_type = "PayByBandwidth"
}

resource "alicloud_ga_endpoint_group" "example" {
  accelerator_id = local.ga_accelerator_id
  endpoint_configurations {
    endpoint = alicloud_eip_address.example.ip_address
    type     = "PublicIp"
    weight   = "20"
  }
  endpoint_group_region = "cn-hangzhou"
  listener_id           = alicloud_ga_listener.example.id
}

resource "alicloud_ga_forwarding_rule" "default" {
  accelerator_id = local.ga_accelerator_id
  listener_id    = alicloud_ga_listener.example.id
  rule_conditions {
    rule_condition_type = "Path"
    path_config {
      values = ["/test"]
    }
  }
  rule_actions {
    order            = "30"
    rule_action_type = "ForwardGroup"
    forward_group_config {
      server_group_tuples {
        endpoint_group_id = alicloud_ga_endpoint_group.example.id
      }
    }
  }
}
`, name)
}
