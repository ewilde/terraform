package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceRunscopeBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketCreate,
		Read:   resourceBucketRead,
		Delete: resourceBucketDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_uuid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBucketCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	log.Printf("[INFO] Creating bucket for name: %s", name)

	bucket, err := createBucketFromResourceData(d)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] bucket create: %#v", bucket)

	result, err := client.CreateBucket(bucket)
	if err != nil {
		return fmt.Errorf("Failed to create bucket: %s", err)
	}

	d.SetId(result)
	log.Printf("[INFO] bucket key: %s", d.Id())

	return resourceBucketRead(d, meta)
}

func resourceBucketRead(d *schema.ResourceData, meta interface{}) error {
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

func resourceBucketDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	key := d.Id()
	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting bucket with key: %s name: %s", key, name)

	if err := client.DeleteBucket(key); err != nil {
		return fmt.Errorf("Error deleting bucket: %s", err)
	}

	return nil
}

func createBucketFromResourceData(d *schema.ResourceData) (Bucket, error) {

	bucket := Bucket{}
	if attr, ok := d.GetOk("name"); ok {
		bucket.Name = attr.(string)
	}
	if attr, ok := d.GetOk("team_uuid"); ok {
		bucket.Team = Team{Id: attr.(string)}
	}

	return bucket, nil
}
