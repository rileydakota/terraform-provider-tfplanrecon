terraform {
  required_providers {
    tfplanrecon = {
      source  = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

data "tfplanrecon_scan" "example" {
    id = "yeet"
    url = "https://your_webhook_here"

}
