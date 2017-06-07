package coredns

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCorednsRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceCorednsRecordCreate,
		Read:   resourceCorednsRecordRead,
		Update: resourceCorednsRecordUpdate,
		Delete: resourceCorednsRecordDelete,

		Schema: map[string]*schema.Schema{
			// Required
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rdata": &schema.Schema{
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Optional
			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			// Computed
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceCorednsRecordCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCorednsRecordRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCorednsRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCorednsRecordDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
