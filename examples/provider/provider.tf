terraform {
  required_providers {
    sonarcloud = {
      source  = "arup-group/sonarcloud"
      version = "0.6.1"
    }
  }
}

provider "sonarcloud" {
  organization = var.organization
  token        = var.token
}
