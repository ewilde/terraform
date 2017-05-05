package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

var _ = fmt.Sprintf("dummy") // dummy
var _ = os.DevNull           // dummy

func TestAccBucket_basic(t *testing.T) {
	var bucketResponse response
	teamId := os.Getenv("RUNSCOPE_TEAM_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testRunscopeBucketConfigA, teamId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("runscope_bucket.test", &bucketResponse),
					resource.TestCheckResourceAttr(
						"runscope_bucket.test", "name", "terraform-provider-test"),
				),
			},
		},
	})
}

func testAccCheckBucketDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "runscope_bucket" {
			continue
		}

		_, err := client.ReadBucket(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Record %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketExists(n string, bucketResponse *response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*Client)

		foundRecord, err := client.ReadBucket(rs.Primary.ID)

		if err != nil {
			return err
		}

		if foundRecord.Data["key"] != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*bucketResponse = foundRecord

		return nil
	}
}

const testRunscopeBucketConfigA = `
resource "runscope_bucket" "bucket" {
  name = "terraform-provider-test"
  team_uuid = "%s"
}`
