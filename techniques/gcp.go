package techniques

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// GcpIamBinding returns the schema for GCP IAM binding creation data source
func GcpIamBinding() *schema.Resource {
	return &schema.Resource{
		ReadContext: gcpIamBindingRead,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The GCP project ID (defaults to GOOGLE_CLOUD_PROJECT env var)",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IAM role to bind (e.g., roles/viewer, roles/editor)",
			},
			"member": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The member to bind the role to (e.g., user:email@example.com, serviceAccount:sa@project.iam.gserviceaccount.com)",
			},
		},
	}
}

func gcpIamBindingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	
	project := d.Get("project").(string)
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if project == "" {
			return diag.FromErr(fmt.Errorf("project must be specified or GOOGLE_CLOUD_PROJECT environment variable must be set"))
		}
	}
	
	role := d.Get("role").(string)
	member := d.Get("member").(string)
	
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "TFPLANRECON GCP IAM Binding Creation",
		Detail:   fmt.Sprintf("Creating IAM binding: Project=%s, Role=%s, Member=%s", project, role, member),
	})
	
	service, err := cloudresourcemanager.NewService(ctx, option.WithScopes(cloudresourcemanager.CloudPlatformScope))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create Cloud Resource Manager service: %v", err))
	}
	
	policy, err := service.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get IAM policy: %v", err))
	}
	
	var binding *cloudresourcemanager.Binding
	for _, b := range policy.Bindings {
		if b.Role == role {
			binding = b
			break
		}
	}
	
	if binding == nil {
		binding = &cloudresourcemanager.Binding{
			Role:    role,
			Members: []string{},
		}
		policy.Bindings = append(policy.Bindings, binding)
	}
	
	for _, existingMember := range binding.Members {
		if existingMember == member {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Member Already Has Role Binding",
				Detail:   fmt.Sprintf("Member %s already has role %s in project %s", member, role, project),
			})
			d.SetId(fmt.Sprintf("%s/%s/%s", project, role, member))
			return diags
		}
	}
	
	binding.Members = append(binding.Members, member)
	
	setRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	
	_, err = service.Projects.SetIamPolicy(project, setRequest).Do()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to set IAM policy: %v", err))
	}
	
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Successfully Added GCP IAM Binding",
		Detail:   fmt.Sprintf("Added %s with role %s to project %s", member, role, project),
	})
	
	d.SetId(fmt.Sprintf("%s/%s/%s", project, role, member))
	return diags
}