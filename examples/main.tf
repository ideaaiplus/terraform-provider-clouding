terraform {
  required_providers {
    clouding = {
      source  = "registry.terraform.io/ideaaiplus/clouding"
      version = "~> 0.1"
    }
  }
}

provider "clouding" {
  api_key = var.clouding_api_key
}

variable "clouding_api_key" {
  type      = string
  sensitive = true
}

resource "clouding_server" "example" {
  name      = "my-terraform-server"
  hostname  = "my-server.example.com"
  image_id  = "ubuntu-22.04-lts"
  flavor_id = "2vcpu-4gb"
  ssh_key_id = "my-ssh-key-id" # optional
}

output "server_id" {
  value = clouding_server.example.id
}
