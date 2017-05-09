package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceRunscopeTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceTestCreate,
		Read:   resourceTestRead,
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
				ForceNew: true,
			},
		},
	}
}

func resourceTestCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	name := d.Get("name").(string)
	log.Printf("[INFO] Creating test with name: %s", name)

	test, err := createTestFromResourceData(d)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] test create: %#v", test)

	createdTest, err := client.CreateTest(test)
	if err != nil {
		return fmt.Errorf("Failed to create test: %s", err)
	}

	d.SetId(createdTest.Id)
	log.Printf("[INFO] test ID: %s", d.Id())

	return resourceTestRead(d, meta)
}

func resourceTestRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	testFromResource, err := createTestFromResourceData(d)
	test, err := client.ReadTest(testFromResource)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Couldn't find test: %s", err)
	}

	d.Set("name", test.Name)
	d.Set("description", test.Description)
	return nil
}

func resourceTestDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	test, err := createTestFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Error deleting test: %s", err)
	}
	log.Printf("[INFO] Deleting test with id: %s name: %s", test.Id, test.Name)

	if err := client.DeleteTest(test); err != nil {
		return fmt.Errorf("Error deleting test: %s", err)
	}

	return nil
}

func createTestFromResourceData(d *schema.ResourceData) (*runscope.Test, error) {

	test := runscope.NewTest()
	test.Id = d.Id()
	if attr, ok := d.GetOk("bucket_id"); ok {
		test.Bucket.Key = attr.(string)
	}

	if attr, ok := d.GetOk("name"); ok {
		test.Name = attr.(string)
	}

	if attr, ok := d.GetOk("description"); ok {
		test.Description = attr.(string)
	}

	return test, nil
}
