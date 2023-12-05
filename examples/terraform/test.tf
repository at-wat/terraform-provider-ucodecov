terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
  required_version = ">= 0.13"
  required_providers {
    ucodecov = {
      source  = "at-wat/ucodecov"
      version = ">= 1.0.0"
    }
  }
}

provider "ucodecov" {
}

variable "owner" {
  default = "your-gh-account"
}

variable "repo" {
  default = "your-repository"
}

data "ucodecov_settings" "test" {
  service = "gh"
  owner   = var.owner
  repo    = var.repo
}

output "test" {
  value = substr(data.ucodecov_settings.test.upload_token, -4, -1)
}
