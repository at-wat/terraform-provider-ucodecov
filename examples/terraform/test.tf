terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
  required_version = ">= 0.13"
  required_providers {
    ucodecov = {
      source  = "at-wat/ucodecov"
      version = "~> 0.0"
    }
  }
}

provider "ucodecov" {
}

data "ucodecov_settings" "test" {
  service = "gh"
  owner   = "your-gh-account"
  repo    = "your-repository"
}

output "test" {
  value = "${data.ucodecov_settings.test.updatestamp}"
}
