package runscope

import (
	"encoding/json"
	"os"
	"testing"
)

var teamId string

func TestCreateBucket(t *testing.T) {
	testPreCheck(t)
	client := clientConfigure()
	key, err := client.CreateBucket(Bucket{Name: "test", Team: Team{Id: teamId}})

	if err != nil {
		t.Error(err)
	}

	client.DeleteBucket(key)
}

func TestReadBucket(t *testing.T) {
	testPreCheck(t)
	client := clientConfigure()

	key, err := client.CreateBucket(Bucket{Name: "terraform-client.go-test", Team: Team{Id: teamId}})
	if err != nil {
		t.Error(err)
	}

	bucket, err := client.ReadBucket(key)
	if err != nil {
		t.Error(err)
	}

	if bucket.Data["key"] != key {
		t.Errorf("Bucket key expected %s was %s.", key, bucket.Data["key"])
	}

	client.DeleteBucket(key)
}

func TestCreateTest(t *testing.T) {
	testPreCheck(t)
	client := clientConfigure()
	key, err := client.CreateBucket(Bucket{Name: "test", Team: Team{Id: teamId}})
	defer client.DeleteBucket(key)

	if err != nil {
		t.Error(err)
	}

	test := Test{Name: "tf_test", Description: "This is a tf test", BucketId: key}
	test, err = client.CreateTest(test)
	defer client.DeleteTest(test)

	if err != nil {
		t.Error(err)
	}

	if len(test.Id) == 0 {
		t.Error("Test id should not be empty")
	}
}

func TestReadTest(t *testing.T) {
	testPreCheck(t)
	client := clientConfigure()
	key, err := client.CreateBucket(Bucket{Name: "test", Team: Team{Id: teamId}})
	defer client.DeleteBucket(key)

	if err != nil {
		t.Error(err)
	}

	test := Test{Name: "tf_test", Description: "This is a tf test", BucketId: key}
	test, err = client.CreateTest(test)
	defer client.DeleteTest(test)

	if err != nil {
		t.Error(err)
	}

	resource, err := client.ReadTest(test)
	if err != nil {
		t.Error(err)
	}

	if resource.Data["name"] != test.Name {
		t.Errorf("Expected name %s, actual %s", test.Name, resource.Data["name"])
	}
}

func TestUpdateTest(t *testing.T) {
	testPreCheck(t)
	client := clientConfigure()
	key, err := client.CreateBucket(Bucket{Name: "test", Team: Team{Id: teamId}})
	defer client.DeleteBucket(key)

	if err != nil {
		t.Error(err)
	}

	test := Test{Name: "tf_test", Description: "This is a tf test", BucketId: key}
	test, err = client.CreateTest(test)
	defer client.DeleteTest(test)

	if err != nil {
		t.Error(err)
	}

	test.Description = "New description"
	resource, err := client.UpdateTest(test)
	if err != nil {
		t.Error(err)
	}

	if resource.Data["description"] != test.Description {
		t.Errorf("Expected description %s, actual %s", test.Description, resource.Data["description"])
	}
}

func TestDeserializeResult(t *testing.T) {
	responseBody := `
	{
	  "meta": {
	    "status": "success"
	  },
	  "data": {
	    "verify_ssl": true,
	    "trigger_url": "https://api.runscope.com/radar/bucket/2e15499d-2e32-4ea8-b6c9-18468031c491/trigger",
	    "name": "foo",
	    "key": "6t0sd3euxlwa",
	    "team": {
	      "name": "form3",
	      "id": "870ed937-bc6e-4d8b-a9a5-d7f9f2412fa3"
	    },
	    "default": false,
	    "auth_token": null,
	    "tests_url": "https://api.runscope.com/buckets/6t0sd3euxlwa/tests",
	    "collections_url": "https://api.runscope.com/buckets/6t0sd3euxlwa/collections",
	    "messages_url": "https://api.runscope.com/buckets/6t0sd3euxlwa/stream"
	  },
	  "error": null
	}
	`
	response := response{}
	err := json.Unmarshal([]byte(responseBody), &response)
	if err != nil {
		t.Error(err)
	}

	if response.Data["key"] != "6t0sd3euxlwa" {
		t.Error("Key not deserialized")
	}
}

func clientConfigure() *Client {
	config := Config{
		AccessToken: os.Getenv("RUNSCOPE_ACCESS_TOKEN"),
		ApiUrl:      "https://api.runscope.com",
	}

	client, _ := config.Client()
	return client
}

func testPreCheck(t *testing.T) {
	skip := os.Getenv("TF_ACC") == ""
	if skip {
		t.Log("runscope client.go tests require setting TF_ACC")
		t.Skip()
	}

	testAccPreCheck(t)
	teamId = os.Getenv("RUNSCOPE_TEAM_ID")
}
