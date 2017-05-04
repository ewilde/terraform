package runscope

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceRunscopeTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceTestCreate,
		Read:   resourceTestRead,
		Update: resourceTestUpdate,
		Delete: resourceTestDelete,

		Schema: map[string]*schema.Schema{
			"bucket_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceTestCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	log.Printf("[INFO] Creating test with name: %s", name)

	test, err := createTestFromResourceData(d)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] test create: %#v", test)

	result, err := client.CreateTest(test)
	if err != nil {
		return fmt.Errorf("Failed to create test: %s", err)
	}

	d.SetId(result.Id)
	log.Printf("[INFO] test ID: %s", d.Id())

	return resourceBucketRead(d, meta)
}

func resourceTestRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	key := d.Id()
	name := d.Get("name").(string)
	log.Printf("[INFO] Reading bucket for id: %s name: %s", key, name)

	bucket, err := client.ReadBucket(key)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Couldn't find bucket: %s", err)
	}

	d.Set("name", bucket.Data["name"])
	d.Set("team_uuid", bucket.Data["team"].(map[string]interface{})["id"])
	return nil
}

func resourceTestUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceTestDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	key := d.Id()
	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting bucket with key: %s name: %s", key, name)

	if err := client.DeleteBucket(key); err != nil {
		return fmt.Errorf("Error deleting bucket: %s", err)
	}

	return nil
}

func createTestFromResourceData(d *schema.ResourceData) (Test, error) {

	test := Test{}
	if attr, ok := d.GetOk("bucket_id"); ok {
		test.BucketId = attr.(string)
	}

	if attr, ok := d.GetOk("name"); ok {
		test.Name = attr.(string)
	}

	if attr, ok := d.GetOk("description"); ok {
		test.Description = attr.(string)
	}

	return test, nil
}
