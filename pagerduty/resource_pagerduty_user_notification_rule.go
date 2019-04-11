package pagerduty

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyUserNotificationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyUserNotificationRuleCreate,
		Read:   resourcePagerDutyUserNotificationRuleRead,
		Update: resourcePagerDutyUserNotificationRuleUpdate,
		Delete: resourcePagerDutyUserNotificationRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyUserNotificationRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateValueFunc([]string{
					"assignment_notification_rule",
				}),
			},

			"start_delay_in_minutes": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"urgency": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateValueFunc([]string{
					"high",
					"low",
				}),
			},
			"contact_method": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"email_contact_method",
								"phone_contact_method",
								"push_notification_contact_method",
								"sms_contact_method",
							}),
						},
					},
				},
			},
		},
	}
}

func buildUserNotificationRuleStruct(d *schema.ResourceData) *pagerduty.NotificationRule {
	var contact_method, err = expandContactMethod(d.Get("contact_method"))
	fmt.Printf("%v\n", contact_method)
	fmt.Printf("%v\n", err)
	notificationRule := &pagerduty.NotificationRule{
		Type:                d.Get("type").(string),
		StartDelayInMinutes: d.Get("start_delay_in_minutes").(int),
		Urgency:             d.Get("urgency").(string),

		/*	ContactMethod: &pagerduty.ContactMethodReference{
					Type: "sms_contact_method",
				ID: "PHE4XPR",
			},
		*/ContactMethod: contact_method,
	}

	return notificationRule
}

func resourcePagerDutyUserNotificationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)

	notificationRule := buildUserNotificationRuleStruct(d)

	resp, _, err := client.Users.CreateNotificationRule(userID, notificationRule)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)

	return resourcePagerDutyUserNotificationRuleRead(d, meta)
}

func resourcePagerDutyUserNotificationRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)

	resp, _, err := client.Users.GetNotificationRule(userID, d.Id())
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("type", resp.Type)
	d.Set("urgency", resp.Urgency)
	d.Set("start_delay_in_minutes", resp.StartDelayInMinutes)
	//	d.Set("contact_method", flattenContactMethod(resp.ContactMethod))

	return nil
}

func resourcePagerDutyUserNotificationRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	contactMethod := buildUserNotificationRuleStruct(d)

	log.Printf("[INFO] Updating PagerDuty user notification rule %s", d.Id())

	userID := d.Get("user_id").(string)

	if _, _, err := client.Users.UpdateNotificationRule(userID, d.Id(), contactMethod); err != nil {
		return err
	}

	return resourcePagerDutyUserNotificationRuleRead(d, meta)
}

func resourcePagerDutyUserNotificationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty user notification rule %s", d.Id())

	userID := d.Get("user_id").(string)

	if _, err := client.Users.DeleteNotificationRule(userID, d.Id()); err != nil {
		return handleNotFoundError(err, d)
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyUserNotificationRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	ids := strings.Split(d.Id(), ":")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_user_notification_rule. Expecting an ID formed as '<user_id>.<notification_rule_id>'")
	}
	uid, id := ids[0], ids[1]

	_, _, err := client.Users.GetNotificationRule(uid, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("user_id", uid)

	return []*schema.ResourceData{d}, nil
}

func expandContactMethod(v interface{}) ([]*pagerduty.ContactMethodReference, error) {
	var contactMethods []*pagerduty.ContactMethodReference

	for _, cm := range v.([]interface{}) {

		rcm := cm.(map[string]interface{})

		contactMethod := &pagerduty.ContactMethodReference{
			ID:   rcm["id"].(string),
			Type: rcm["type"].(string),
		}

		contactMethods = append(contactMethods, contactMethod)
	}

	return contactMethods, nil
}

func flattenContactMethod(v []*pagerduty.ContactMethod) []map[string]interface{} {
	var contactMethods []map[string]interface{}

	for _, cm := range v {

		contactMethod := map[string]interface{}{
			"id":   cm.ID,
			"type": cm.Type,
		}

		contactMethods = append(contactMethods, contactMethod)
	}

	return contactMethods
}
