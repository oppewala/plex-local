variable "azure_group_name" {
  type        = string
  description = "Name that will be used to generate resource group and resources"
  sensitive   = true
}


module "func" {
  source = "./modules/azure"

  group_name = var.azure_group_name
}

#module "local" {
#  source = "./modules/local"
#}