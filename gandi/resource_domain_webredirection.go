package gandi

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-gandi/go-gandi/domain"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDomainWebRedirection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainWebRedirectionCreate,
		Read:          resourceDomainWebRedirectionRead,
		Update:        resourceDomainWebRedirectionUpdate,
		Delete:        resourceDomainWebRedirectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The domain name",
			},
			"host": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Source hostname (including the domain name)",
				ValidateDiagFunc: validateDomainWebRedirectionHostname,
			},
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Target URL",
				ValidateDiagFunc: validateDomainWebRedirectionURL,
			},
			"override": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When you create a redirection on a domain a DNS record is created if it does not exist. When the record already exists and this parameter is set to true it will overwrite the record. Otherwise it will trigger an error.",
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Protocol of the redirection",
				Default:          "https",
				ValidateDiagFunc: valitateDomainWebRedirectionProtocol,
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Type of redirection",
				Default:          "http301",
				ValidateDiagFunc: valitateDomainWebRedirectionType,
			},
			"cert_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the TLS certificate used for HTTPS redirection (none, pending, active, error)",
			},
			"cert_uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of the TLS certificate used for HTTPS redirection",
			},
		},
	}
}

func resourceDomainWebRedirectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients).Domain

	var err error
	hostname := d.Get("host").(string)
	domainname := d.Get("domain").(string)
	if domainname == "" {
		domainname, err = extractDomain(hostname)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if !strings.HasSuffix(hostname, domainname) {
		return diag.FromErr(fmt.Errorf("the hostname %q does not end with the domain name %q", hostname, domainname))
	}

	request := domain.WebRedirectionCreateRequest{
		Host:     hostname,
		Override: d.Get("override").(bool),
		Protocol: d.Get("protocol").(string),
		Type:     d.Get("type").(string),
		URL:      d.Get("url").(string),
	}

	err = client.CreateWebRedirection(domainname, request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create web redirection for %s: %w", d.Id(), err))
	}

	d.SetId(hostname)

	return diag.FromErr(resourceDomainWebRedirectionRead(d, meta))
}

func resourceDomainWebRedirectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients).Domain

	var err error
	hostname := d.Id()
	domainname := d.Get("domain").(string)
	if domainname == "" {
		domainname, err = extractDomain(hostname)
		if err != nil {
			return fmt.Errorf("failed to extract domain from hostname %q: %w", hostname, err)
		}
	}
	if !strings.HasSuffix(hostname, domainname) {
		return fmt.Errorf("the hostname %q does not end with the domain name %q", hostname, domainname)
	}

	response, err := client.GetWebRedirection(domainname, hostname)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("failed to get web redirection data for %s: %w", hostname, err)
	}

	d.SetId(response.Host)
	if err = d.Set("domain", domainname); err != nil {
		return fmt.Errorf("failed to set domain for %s: %w", d.Id(), err)
	}
	if err = d.Set("host", response.Host); err != nil {
		return fmt.Errorf("failed to set host for %s: %w", d.Id(), err)
	}
	if err = d.Set("url", response.URL); err != nil {
		return fmt.Errorf("failed to set url for %s: %w", d.Id(), err)
	}
	if err = d.Set("protocol", response.Protocol); err != nil {
		return fmt.Errorf("failed to set protocol for %s: %w", d.Id(), err)
	}
	if err = d.Set("type", response.Type); err != nil {
		return fmt.Errorf("failed to set type for %s: %w", d.Id(), err)
	}
	if err = d.Set("cert_status", response.CertificateStatus); err != nil {
		return fmt.Errorf("failed to set cert_status for %s: %w", d.Id(), err)
	}
	if err = d.Set("cert_uuid", response.CertificateUUID); err != nil {
		return fmt.Errorf("failed to set cert_uuid for %s: %w", d.Id(), err)
	}

	return nil
}

func resourceDomainWebRedirectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients).Domain

	domainname := d.Get("domain").(string)
	hostname := d.Id()

	request := domain.WebRedirectionUpdateRequest{
		Override: d.Get("override").(bool),
		Protocol: d.Get("protocol").(string),
		Type:     d.Get("type").(string),
		URL:      d.Get("url").(string),
	}

	err := client.UpdateWebRedirection(domainname, hostname, request)
	if err != nil {
		return fmt.Errorf("failed to update web redirection for %s: %w", d.Id(), err)
	}

	return resourceDomainWebRedirectionRead(d, meta)
}

func resourceDomainWebRedirectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients).Domain

	domainname := d.Get("domain").(string)
	hostname := d.Id()

	if err := client.DeleteWebRedirection(domainname, hostname); err != nil {
		return fmt.Errorf("failed to delete web redirection for %s: %w", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func extractDomain(fqdn string) (string, error) {
	// A domain is composed of a name and a TLD. The TLD can be
	// composed of several parts (e.g. co.uk). We consider that
	// the domain name is the two last parts of the FQDN.
	parts := strings.Split(fqdn, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("%q is not a valid domain name", fqdn)
	}
	domain := fmt.Sprintf("%s.%s", parts[len(parts)-2], parts[len(parts)-1])
	return domain, nil
}

func validateDomainWebRedirectionHostname(v interface{}, p cty.Path) diag.Diagnostics {
	value := v.(string)

	// Hostname regex pattern: allows alphanumeric characters, hyphens, and dots
	// Must start and end with alphanumeric character, no consecutive dots
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

	if hostnameRegex.MatchString(value) {
		return nil
	}

	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  "Invalid hostname",
			Detail:   fmt.Sprintf("%q is not a valid hostname", value),
		},
	}
}

func validateDomainWebRedirectionURL(v interface{}, p cty.Path) diag.Diagnostics {
	value := v.(string)

	// Basic URL regex pattern
	urlRegex := regexp.MustCompile(`^(https?://)?([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}(/.*)?$`)

	if urlRegex.MatchString(value) {
		return nil
	}

	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  "Invalid URL",
			Detail:   fmt.Sprintf("%q is not a valid URL", value),
		},
	}
}

func valitateDomainWebRedirectionType(v interface{}, p cty.Path) diag.Diagnostics {
	value := v.(string)
	validTypes := []string{"cloak", "http301", "http302"}

	return valitateDomainWebRedirectionEnum("type", value, validTypes)
}

func valitateDomainWebRedirectionProtocol(v interface{}, p cty.Path) diag.Diagnostics {
	value := v.(string)
	validProtocols := []string{"http", "https", "httpsonly"}

	return valitateDomainWebRedirectionEnum("protocol", value, validProtocols)
}

func valitateDomainWebRedirectionEnum(name string, value string, enumValues []string) (diags diag.Diagnostics) {
	for _, acceptedValue := range enumValues {
		if value == acceptedValue {
			return
		}
	}

	diag := diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "wrong value",
		Detail:   fmt.Sprintf("%q is not a valid %s", value, name),
	}
	diags = append(diags, diag)

	return
}
