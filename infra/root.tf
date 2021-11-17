terraform {
  backend "remote" {
    organization = "crackedjar"

    workspaces {
      name = "plex-local"
    }
  }
}


module "func" {
  source = "./modules/azure"

  group_name = var.azure_group_name
}

#module "local" {
#  source = "./modules/local"
#}