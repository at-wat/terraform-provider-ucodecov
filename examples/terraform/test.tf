terraform {
  backend "local" {
    path = "terraform.tfstate"
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
