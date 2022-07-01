package cloudstack

import (
	"fmt"
	"testing"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudStackVPC_basic(t *testing.T) {
	var vpc cloudstack.VPC

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackVPC_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackVPCExists(
						"cloudstack_vpc.foo", &vpc),
					testAccCheckCloudStackVPCAttributes(&vpc),
					resource.TestCheckResourceAttr(
						"cloudstack_vpc.foo", "vpc_offering", "Default VPC offering"),
					testAccCheckResourceTags(&vpc),
				),
			},
		},
	})
}

func TestAccCloudStackVPC_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackVPC_basic,
			},

			{
				ResourceName:      "cloudstack_vpc.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudStackVPCExists(
	n string, vpc *cloudstack.VPC) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		v, _, err := cs.VPC.GetVPCByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if v.Id != rs.Primary.ID {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *v

		return nil
	}
}

func testAccCheckCloudStackVPCAttributes(
	vpc *cloudstack.VPC) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if vpc.Name != "terraform-vpc" {
			return fmt.Errorf("Bad name: %s", vpc.Name)
		}

		if vpc.Displaytext != "terraform-vpc-text" {
			return fmt.Errorf("Bad display text: %s", vpc.Displaytext)
		}

		if vpc.Cidr != "10.0.0.0/8" {
			return fmt.Errorf("Bad VPC CIDR: %s", vpc.Cidr)
		}

		if vpc.Networkdomain != "terraform-domain" {
			return fmt.Errorf("Bad network domain: %s", vpc.Networkdomain)
		}

		return nil
	}
}

func testAccCheckCloudStackVPCDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_vpc" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		_, _, err := cs.VPC.GetVPCByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("VPC %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCloudStackVPC_basic = `
resource "cloudstack_vpc" "foo" {
  name = "terraform-vpc"
  display_text = "terraform-vpc-text"
  cidr = "10.0.0.0/8"
  vpc_offering = "Default VPC offering"
  network_domain = "terraform-domain"
  zone = "Sandbox-simulator"
  tags = {
    terraform-tag = "true"
  }
}`
