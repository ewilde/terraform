package runscope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
	"io/ioutil"
	"log"
	"strings"
)

type Client struct {
	ApiUrl      string
	AccessToken string
	Http        *http.Client
}

type response struct {
	Meta  metaResponse           `json:"meta"`
	Data  map[string]interface{} `json:"data"`
	Error errorResponse          `json:"error"`
}

type collectionResponse struct {
	Meta  metaResponse  `json:"meta"`
	Data  []interface{} `json:"data"`
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"error"`
}

type metaResponse struct {
	Status string `json:"status"`
}

func NewClient(apiUrl string, accessToken string) *Client {
	client := Client{
		ApiUrl:      apiUrl,
		AccessToken: accessToken,
		Http:        cleanhttp.DefaultClient(),
	}

	return &client
}

type Bucket struct {
	Id   string
	Name string
	Team Team
}

type Team struct {
	Name string
	Id   string
}

func (client *Client) CreateBucket(bucket Bucket) (string, error) {

	data := url.Values{}
	data.Add("name", bucket.Name)
	data.Add("team_uuid", bucket.Team.Id)

	req, err := client.newFormUrlEncodedRequest("POST", "/buckets", data)
	if err != nil {
		return "", err
	}

	resp, err := client.Http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Printf("[DEBUG] %s", bodyString)

	if resp.StatusCode >= 300 {
		errorResp := new(errorResponse)
		if err = json.Unmarshal(bodyBytes, &errorResp); err != nil {
			return "", fmt.Errorf("Error creating bucket: %s", bucket.Name)
		} else {
			return "", fmt.Errorf("Error creating bucket: %s, status: %d reason: %q", bucket.Name,
				errorResp.Status, errorResp.ErrorMessage)
		}
	} else {
		response := new(response)
		json.Unmarshal(bodyBytes, &response)
		return response.Data["key"].(string), nil
	}
}

func (client *Client) ReadBucket(key string) (response, error) {
	response := response{}

	req, err := client.newRequest("GET", fmt.Sprintf("/buckets/%s", key), nil)
	if err != nil {
		return response, err
	}

	resp, err := client.Http.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Printf("[DEBUG] %s", bodyString)

	if resp.StatusCode >= 300 {
		errorResp := new(errorResponse)
		if err = json.Unmarshal(bodyBytes, &errorResp); err != nil {
			return response, fmt.Errorf("Status: %s Error reading bucket: %s",
				resp.Status, key)
		} else {
			return response, fmt.Errorf("Status: %s Error reading bucket: %s, reason: %q",
				resp.Status, key, errorResp.ErrorMessage)
		}
	} else {
		json.Unmarshal(bodyBytes, &response)
		return response, nil
	}
}

func (client *Client) DeleteBucket(key string) error {

	req, err := client.newRequest("DELETE", fmt.Sprintf("/buckets/%s", key), nil)
	if err != nil {
		return err
	}

	resp, err := client.Http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		errorResp := new(errorResponse)

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Printf("[DEBUG] %s", bodyString)

		if err = json.Unmarshal(bodyBytes, &errorResp); err != nil {
			return fmt.Errorf("Error creating bucket: %s", key)
		} else {
			return fmt.Errorf("Error creating bucket: %s, status: %d reason: %q", key,
				errorResp.Status, errorResp.ErrorMessage)
		}
	}

	return nil
}

func (client *Client) newFormUrlEncodedRequest(method string, endpoint string, data url.Values) (*http.Request, error) {

	var urlStr string
	urlStr = client.ApiUrl + endpoint
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("Error during parsing request URL: %s", err)
	}

	req, err := http.NewRequest(method, url.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Error during creation of request: %s", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.AccessToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (client *Client) newRequest(method string, endpoint string, body []byte) (*http.Request, error) {

	var urlStr string
	urlStr = client.ApiUrl + endpoint
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("Error during parsing request URL: %s", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of request: %s", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.AccessToken))
	req.Header.Add("Accept", "application/json")

	if method != "GET" {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}
