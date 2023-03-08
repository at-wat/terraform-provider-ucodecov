# Codecov Provider (unofficial)

This provider presents Codecov repository data.
It's mainly for getting the coverage upload token.

- API reference: https://docs.codecov.com/reference/overview

## Example Usage

Example to set upload token to GitHub Actions as a secret.
```hcl
terraform {
  required_providers {
    ucodecov = {
      source = "at-wat/ucodecov"
    }
  }
}

locals {
  owner = "at-wat"
  repo  = "terraform-provider-ucodecov"
}


provider "ucodecov" {
}

provider "github" {
  organization = local.owner
  version      = "~> 2.6"
}


data "ucodecov_settings" "this" {
  service = "gh"
  owner   = local.owner
  repo    = local.repo
}

resource "github_actions_secret" "example" {
  repository      = local.repo
  secret_name     = "CODECOV_TOKEN"
  plaintext_value = data.ucodecov_settings.this.upload_token
}
```
