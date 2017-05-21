package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

func TestAccEnvironment_basic(t *testing.T) {
	var environment runscope.Environment
	teamId := os.Getenv("RUNSCOPE_TEAM_ID")

	testCheck := func(*terraform.State) error {
		if len(environment.Integrations) != 2 {
			return fmt.Errorf("Expected %d integrations, actual %d", 2, len(environment.Integrations))
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testRunscopeEnvrionmentConfigA, teamId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists("runscope_environment.environment", &environment),
					testCheck,
					resource.TestCheckResourceAttr(
						"runscope_environment.environment", "name", "test-environment")),

			},
		},
	})
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*runscope.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "runscope_environment" {
			continue
		}

		var err error
		bucketId := rs.Primary.Attributes["bucket_id"]
		testId := rs.Primary.Attributes["test_id"]
		if testId != "" {
			err = client.DeleteTestEnvironment(&runscope.Environment{ID: rs.Primary.ID},
				&runscope.Test{
					ID:     testId,
					Bucket: &runscope.Bucket{Key: bucketId}})
		} else {
			err = client.DeleteSharedEnvironment(&runscope.Environment{ID: rs.Primary.ID},
				&runscope.Bucket{Key: bucketId})
		}

		if err == nil {
			return fmt.Errorf("Record %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEnvironmentExists(n string, environment *runscope.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*runscope.Client)

		var foundRecord *runscope.Environment
		var err error

		environment.ID = rs.Primary.ID
		bucketId := rs.Primary.Attributes["bucket_id"]
		testId := rs.Primary.Attributes["test_id"]
		if testId != "" {
			foundRecord, err = client.ReadTestEnvironment(environment,
				&runscope.Test{
					ID:     testId,
					Bucket: &runscope.Bucket{Key: bucketId}})
		} else {
			foundRecord, err = client.ReadSharedEnvironment(environment,
				&runscope.Bucket{Key: bucketId})
		}

		if err != nil {
			return err
		}

		if foundRecord.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		environment = foundRecord
		return nil
	}
}

const testRunscopeEnvrionmentConfigA = `
resource "runscope_environment" "environment" {
  bucket_id    = "${runscope_bucket.bucket.id}"
  name         = "test-environment"

  integrations = [
    {
      integration_type = "pagerduty"
      description      = "alert on call"
    },
    {
      integration_type = "slack"
      description      = "post to slack"
    }
  ]

  initial_variables {
    var1 = "true",
    var2 = "value2"
  }
}

resource "runscope_test" "test" {
  bucket_id = "${runscope_bucket.bucket.id}"
  name = "runscope test"
  description = "This is a test test..."
}

resource "runscope_bucket" "bucket" {
  name = "terraform-provider-test"
  team_uuid = "%s"
}
`
