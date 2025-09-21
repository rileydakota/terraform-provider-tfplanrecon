package main

import (
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rileydakota/tf-plan-recon/techniques"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"tfplanrecon_env_var_exfil":   techniques.EnvVarExfil(),
			"tfplanrecon_env_var_print":   techniques.EnvVarPrint(),
			"tfplanrecon_gcp_iam_binding": techniques.GcpIamBinding(),
			"tfplanrecon_aws_iam_role":    techniques.AwsIamRole(),
			"tfplanrecon_aws_secrets":     techniques.AwsSecretsExfil(),
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

	return &techniques.ProviderConfig{
		Client: client,
	}, nil
}