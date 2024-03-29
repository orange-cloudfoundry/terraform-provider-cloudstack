package cloudstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudStackVPNGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackVPNGatewayCreate,
		Read:   resourceCloudStackVPNGatewayRead,
		Delete: resourceCloudStackVPNGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudStackVPNGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	vpcid := d.Get("vpc_id").(string)
	p := cs.VPN.NewCreateVpnGatewayParams(vpcid)

	// Create the new VPN Gateway
	v, err := cs.VPN.CreateVpnGateway(p)
	if err != nil {
		return fmt.Errorf("Error creating VPN Gateway for VPC ID %s: %s", vpcid, err)
	}

	d.SetId(v.Id)

	return resourceCloudStackVPNGatewayRead(d, meta)
}

func resourceCloudStackVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Get the VPN Gateway details
	v, count, err := cs.VPN.GetVpnGatewayByID(d.Id())
	if err != nil {
		if count == 0 {
			log.Printf(
				"[DEBUG] VPN Gateway for VPC ID %s does no longer exist", d.Get("vpc_id").(string))
			d.SetId("")
			return nil
		}

		return err
	}

	if err = d.Set("vpc_id", v.Vpcid); err != nil {
		return err
	}
	if err = d.Set("public_ip", v.Publicip); err != nil {
		return err
	}

	return nil
}

func resourceCloudStackVPNGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Create a new parameter struct
	p := cs.VPN.NewDeleteVpnGatewayParams(d.Id())

	// Delete the VPN Gateway
	_, err := cs.VPN.DeleteVpnGateway(p)
	if err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", d.Id())) {
			return nil
		}

		return fmt.Errorf("Error deleting VPN Gateway for VPC %s: %s", d.Get("vpc_id").(string), err)
	}

	return nil
}
