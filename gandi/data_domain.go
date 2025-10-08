package gandi

import (
	"fmt"

	"github.com/go-gandi/go-gandi/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The FQDN of the domain",
			},
			"nameservers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of nameservers for the domain",
			},
		},
		Read: dataSourceDomainRead,
	}
}

func dataSourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients).Domain

	name := d.Get("name").(string)

	found, err := client.GetDomain(name)

	if err != nil {
		requestError, ok := err.(*types.RequestError)
		if ok && requestError.StatusCode == 404 {
			return fmt.Errorf("unknown domain '%s': %w", name, err)
		}

		return fmt.Errorf("failed to retrieve domain '%s' data: %w", name, err)
	}

	d.SetId(found.FQDN)
	if err = d.Set("name", found.FQDN); err != nil {
		return fmt.Errorf("failed to set name for %s: %w", d.Id(), err)
	}

	if err = d.Set("nameservers", found.Nameservers); err != nil {
		return fmt.Errorf("failed to set nameservers for %s: %w", d.Id(), err)
	}

	return nil
}
