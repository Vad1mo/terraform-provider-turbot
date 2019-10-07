package turbot

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-turbot/apiClient"
)

func resourceTurbotPolicySetting() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotPolicySettingCreate,
		Read:   resourceTurbotPolicySettingRead,
		Update: resourceTurbotPolicySettingUpdate,
		Delete: resourceTurbotPolicySettingDelete,
		Exists: resourceTurbotPolicySettingExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotPolicySettingImport,
		},
		Schema: map[string]*schema.Schema{
			"policy_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressIfEncryptedOrValueSourceMatches,
			},
			"value_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value_source_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"precedence": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "required",
			},
			"template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"template_input": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"note": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_from_timestamp": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_to_timestamp": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"value_source_used": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceTurbotPolicySettingExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*apiClient.Client)
	id := d.Id()

	_, err := client.ReadPolicySetting(id)
	if err != nil {
		if apiClient.NotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceTurbotPolicySettingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	policyTypeUri := d.Get("policy_type").(string)
	resourceAka := d.Get("resource").(string)

	// first check if the folder exists - search by parent and foldere title
	existingSetting, err := client.FindPolicySetting(policyTypeUri, resourceAka)
	if err != nil {
		return err
	}
	if existingSetting.Value != nil {
		return fmt.Errorf("A policy setting for policy type: '%s', resource: '%s' already exists ( id: %s ). To manage the existing setting using Terraform, import it using command 'terraform import <resource_address> <id>'",
			policyTypeUri, resourceAka, existingSetting.Turbot.Id)
	}

	// NOTE:  turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.
	//	as we are not sure of the value format provided, we try multiple times, relying on the policy validation to
	//	reject invalid values
	// 1) pass value as 'value'
	// 2) pass value as 'valueSource'. update d.value to be the yaml parsed version of 'value'
	commandPayload := buildPayload(d)
	setting, err := client.CreatePolicySetting(policyTypeUri, resourceAka, commandPayload)
	if err != nil {
		if !apiClient.FailedValidationError(err) {
			d.SetId("")
			return err
		}
		// so we have a data validation error, try the value source
		commandPayload["valueSource"] = commandPayload["value"]
		delete(commandPayload, "value")
		// try again
		setting, err = client.CreatePolicySetting(policyTypeUri, resourceAka, commandPayload)
		if err != nil {
			d.SetId("")
			return err
		}
		// update state value setting with yaml parsed valueSource
		setValueFromValueSource(commandPayload["valueSource"], d)
	}
	// if pgp_key has been supplied, encrypt value and value_source
	storeValue(d, setting)

	// assign the id
	d.SetId(setting.Turbot.Id)

	return nil
}

func resourceTurbotPolicySettingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	setting, err := client.ReadPolicySetting(id)
	if err != nil {
		if apiClient.NotFoundError(err) {
			// setting was not found - clear id
			d.SetId("")
		}
		return err
	}

	// assign results back into ResourceData

	// if pgp_key has been supplied, encrypt value and value_source
	storeValue(d, setting)
	d.Set("precedence", setting.Precedence)
	d.Set("template", setting.Template)
	d.Set("template_input", setting.TemplateInput)
	d.Set("note", setting.Note)
	d.Set("valid_from_timestamp", setting.ValidFromTimestamp)
	d.Set("valid_to_timestamp", setting.ValidToTimestamp)

	return nil
}

func resourceTurbotPolicySettingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	// NOTE:  turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.
	//	as we are not sure of the value format provided, we try multiple times, relying on the policy validation to
	//	reject invalid values
	// 1) pass value as 'value'
	// 2) pass value as 'valueSource'. update d.value to be the yaml parsed version of 'value'
	commandPayload := buildPayload(d)

	err := client.UpdatePolicySetting(id, commandPayload)
	if err != nil {
		if !apiClient.FailedValidationError(err) {
			d.SetId("")
			return err
		}
		// so we have a data validation error - try using value as valueSource
		commandPayload["valueSource"] = commandPayload["value"]
		delete(commandPayload, "value")
		// try again
		err := client.UpdatePolicySetting(id, commandPayload)
		if err != nil {
			d.SetId("")
			return err
		}
		// update state value setting with yaml parsed valueSource
		setValueFromValueSource(commandPayload["valueSource"], d)
	}

	return nil
}

func setValueFromValueSource(valueSource string, d *schema.ResourceData) {
	var i interface{}
	yaml.Unmarshal([]byte(valueSource), &i)
	d.Set("value", fmt.Sprintf("%v", i))
	d.Set("value_source", valueSource)
	d.Set("value_source_used", true)
}

func resourceTurbotPolicySettingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()
	err := client.DeletePolicySetting(id)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceTurbotPolicySettingImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotPolicySettingRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildPayload(d *schema.ResourceData) map[string]string {
	commandPayload := map[string]string{
		"value":              d.Get("value").(string),
		"precedence":         d.Get("precedence").(string),
		"template":           d.Get("template").(string),
		"templateInput":      d.Get("template_input").(string),
		"note":               d.Get("note").(string),
		"validFromTimestamp": d.Get("valid_from_timestamp").(string),
		"validToTimestamp":   d.Get("valid_to_timestamp").(string),
	}
	// remove nil entries from commandPayload
	for k, v := range commandPayload {
		if v == "" {
			delete(commandPayload, k)
		}
	}
	return commandPayload
}

// If a pgp key is present, value will be encrypted so we cannot perform diff
// If valueSource was used, suppress diff if value source matches
func suppressIfEncryptedOrValueSourceMatches(_, old, new string, d *schema.ResourceData) bool {
	// if old value is not set, do not suppress - cannot be encrypted and value source will not have been used
	if old == "" {
		return false
	}

	_, keyPresent := d.GetOk("pgp_key")

	// Return true if the diff should be suppressed, false to retain it.
	if d.Get("value_source_used").(bool) {
		old = d.Get("value_source").(string)
	}
	return keyPresent || new == old
}

// write value and value_source to ResourceData, encrypting if a pgp key was provided
func storeValue(d *schema.ResourceData, setting *apiClient.PolicySetting) error {
	// NOTE: turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.

	if pgpKey, ok := d.GetOk("pgp_key"); ok {
		// format the value as a string to allow us to handle object/array values using a string schema
		valueFingerprint, encryptedValue, err := encryptValue(pgpKey.(string), settingValueToString(setting.Value))
		if err != nil {
			return err
		}
		d.Set("value", encryptedValue)
		d.Set("value_key_fingerprint", valueFingerprint)

		valueSourceFingerprint, encryptedValueSource, err := encryptValue(pgpKey.(string), setting.ValueSource)
		if err != nil {
			return err
		}
		d.Set("value_source", encryptedValueSource)
		d.Set("value_source_key_fingerprint", valueSourceFingerprint)
	} else {
		d.Set("value", settingValueToString(setting.Value))
		d.Set("value_source", setting.ValueSource)
	}

	return nil
}

// convert value to a string. If it is a complex type (object/array) then the diff calculation will use the value_source so the precise format is not critical
func settingValueToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
