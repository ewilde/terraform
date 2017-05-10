package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceRunscopeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentCreate,
		Read:   resourceEnvironmentRead,
		Delete: resourceEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"bucket_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"test_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	name := d.Get("name").(string)
	log.Printf("[INFO] Creating environment with name: %s", name)

	environment, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] environment create: %#v", environment)

	var createdEnvironment *runscope.Environment
	bucket_id := d.Get("bucket_id").(string)
	test_id := d.Get("test_id").(string)
	if sharedEnvironment(d) {
		createdEnvironment, err = client.CreateTestEnvironment(environment,
			&runscope.Test{Id: test_id, Bucket: &runscope.Bucket{Key: bucket_id}})
	} else {
		createdEnvironment, err = client.CreateSharedEnvironment(environment,
			&runscope.Bucket{Key: bucket_id})
	}
	if err != nil {
		return fmt.Errorf("Failed to create environment: %s", err)
	}

	d.SetId(createdEnvironment.Id)
	log.Printf("[INFO] environment ID: %s", d.Id())

	return resourceEnvironmentRead(d, meta)
}

func resourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	environmentFromResource, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Failed to read environment from resource data: %s", err)
	}

	var environment *runscope.Environment
	bucket_id := d.Get("bucket_id").(string)
	test_id := d.Get("test_id").(string)

	if sharedEnvironment(d) {
		environment, err = client.ReadSharedEnvironment(
			environmentFromResource, &runscope.Bucket{Key: bucket_id})
	} else {
		environment, err = client.ReadTestEnvironment(
			environmentFromResource, &runscope.Test{Id: test_id, Bucket: &runscope.Bucket{Key: bucket_id}})
	}

	if err != nil {
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Couldn't find environment: %s", err)
	}

	d.Set("name", environment.Name)
	return nil
}

func resourceEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	environmentFromResource, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Failed to read environment from resource data: %s", err)
	}

	var environment *runscope.Environment
	bucketId := d.Get("bucket_id").(string)
	testId := d.Get("test_id").(string)

	if sharedEnvironment(d) {
		log.Printf("[INFO] Deleting shared environment with id: %s name: %s",
			environment.Id, environment.Name)
		environment, err = client.ReadSharedEnvironment(
			environmentFromResource, &runscope.Bucket{Key: bucketId})
	} else {
		log.Printf("[INFO] Deleting shared environment with id: %s name: %s, from test %s",
			environment.Id, environment.Name, testId)
		environment, err = client.ReadTestEnvironment(
			environmentFromResource, &runscope.Test{Id: testId, Bucket: &runscope.Bucket{Key: bucketId}})
	}

	if err != nil {
		return fmt.Errorf("Error deleting environment: %s", err)
	}

	return nil
}

func createEnvironmentFromResourceData(d *schema.ResourceData) (*runscope.Environment, error) {

	environment := runscope.NewEnvironment()
	environment.Id = d.Id()

	if attr, ok := d.GetOk("name"); ok {
		environment.Name = attr.(string)
	}

	return environment, nil
}

func sharedEnvironment(d *schema.ResourceData) bool {
	return len(d.Get("bucket_id").(string)) > 0 &&
		len(d.Get("test_id").(string)) > 0
}
