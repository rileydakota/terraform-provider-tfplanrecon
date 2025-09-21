# Environment Variable Console Printing Example
# This data source will print all environment variables to the console

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Print environment variables in plain text
data "tfplanrecon_env_var_print" "plain" {
  base64_encode = false
}

# Print environment variables base64 encoded
data "tfplanrecon_env_var_print" "encoded" {
  base64_encode = true
}