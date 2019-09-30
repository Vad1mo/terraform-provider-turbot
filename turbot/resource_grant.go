package turbot

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-turbot/apiclient"
)

// map of Terraform properties to Turbot properties that we pass to create and update mutations
var grantDataMap = map[string]string{
	"permission_type":  "permissionTypeAka",
	"permission_level": "permissionLevelAka",
}

func resourceTurbotGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotGrantCreate,
		Read:   resourceTurbotGrantRead,
		Delete: resourceTurbotGrantDelete,
		Exists: resourceTurbotGrantExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotGrantImport,
		},
		Schema: map[string]*schema.Schema{
			// aka of the resource resource
			"resource": {
				Type:     schema.TypeString,
				Required: true,
				// when doing a diff, the state file will contain the id of the resource bu tthe config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: supressIfResourceAkaMatches,
				ForceNew:         true,
			},
			// when doing a read, fetch the resource akas to use in supressIfResourceAkaMatches()
			"resource_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},
			"permission_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permission_level": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"profile_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTurbotGrantExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*apiclient.Client)
	id := d.Id()
	return client.GrantExists(id)
}

func resourceTurbotGrantCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiclient.Client)
	resourceAka := d.Get("resource").(string)
	profileId := d.Get("profile_id").(string)
	// build map of Grant properties
	data := mapFromResourceDataWithPropertyMap(d, grantDataMap)
	// create Grant returns turbot resource metadata containing the id
	TurbotGrantMetadata, err := client.CreateGrant(profileId, resourceAka, data)
	if err != nil {
		return err
	}

	// set resource_akas property by loading resource resource and fetching the akas
	resource_akas, err := client.GetResourceAkas(resourceAka)
	if err != nil {
		return err
	}
	// assign resource akas
	d.Set("resource_akas", resource_akas)

	// assign the id
	d.SetId(TurbotGrantMetadata.Id)
	return nil
}

func resourceTurbotGrantRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiclient.Client)
	id := d.Id()

	Grant, err := client.ReadGrant(id)
	if err != nil {
		if apiclient.NotFoundError(err) {
			// Grant was not found - clear id
			d.SetId("")
		}
		return err
	}

	// assign results back into ResourceData

	// set resource_akas property by loading resource resource and fetching the akas
	resource_akas, err := client.GetResourceAkas(Grant.Turbot.ResourceId)
	if err != nil {
		return err
	}
	d.Set("permission_level_id", Grant.PermissionLevelId)
	d.Set("permission_type_id", Grant.PermissionTypeId)
	d.Set("profile_id", Grant.Turbot.ProfileId)
	d.Set("resource", Grant.Turbot.ResourceId)
	d.Set("resource_akas", resource_akas)
	return nil
}

func resourceTurbotGrantDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiclient.Client)
	id := d.Id()
	err := client.DeleteGrant(id)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceTurbotGrantImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotGrantRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func supressIfResourceAkaMatches(k, old, new string, d *schema.ResourceData) bool {
	resourceAkasProperty, resourceAkasSet := d.GetOk("resource_akas")
	// if resource_id has not been set yet, do not suppress the diff
	if !resourceAkasSet {
		return false
	}

	resource_Akas, ok := resourceAkasProperty.([]interface{})
	if !ok {
		return false
	}
	// if resource_akas contains the 'new' aka, suppress diff
	for _, aka := range resource_Akas {
		if aka.(string) == new {
			return true
		}
	}
	return false
}