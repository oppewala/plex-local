variable "group_name" {
  default   = ""
  type      = string
  sensitive = true
}

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.65"
    }
  }

  required_version = ">= 0.14.9"
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

locals {
  safe_group_name = replace(var.group_name, "-", "")
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${var.group_name}"
  location = "Australia East"
}

resource "azurerm_application_insights" "core" {
  name                = "ai-${var.group_name}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  application_type    = "web"
}

resource "azurerm_key_vault" "core" {
  name                = "kv-${var.group_name}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  sku_name            = "standard"
  tenant_id           = data.azurerm_client_config.current.tenant_id
}

# Function App
resource "azurerm_storage_account" "fapp" {
  name                     = "safapp${local.safe_group_name}"
  location                 = azurerm_resource_group.rg.location
  resource_group_name      = azurerm_resource_group.rg.name
  account_replication_type = "LRS"
  account_tier             = "Standard"
}

resource "azurerm_app_service_plan" "fapp" {
  name                = "asp-plex-local-dl"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  kind                = "FunctionApp"

  sku {
    size = "Y1"
    tier = "Dynamic"
  }
}

resource "azurerm_function_app" "fapp" {
  name                       = "fapp-${var.group_name}"
  location                   = azurerm_resource_group.rg.location
  resource_group_name        = azurerm_resource_group.rg.name
  app_service_plan_id        = azurerm_app_service_plan.fapp.id
  storage_account_name       = azurerm_storage_account.fapp.name
  storage_account_access_key = azurerm_storage_account.fapp.primary_access_key
  https_only                 = true
  version                    = "~4"
  app_settings               = {
    "WEBSITE_RUN_FROM_PACKAGE"       = ""
    "FUNCTIONS_WORKER_RUNTIME"       = "node"
    "WEBSITE_NODE_DEFAULT_VERSION"   = "10.14.1"
    "APPINSIGHTS_INSTRUMENTATIONKEY" = azurerm_application_insights.core.instrumentation_key
  }
  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_key_vault_access_policy" "fapp" {
  key_vault_id = azurerm_key_vault.core.id
  object_id    = azurerm_function_app.fapp.identity.principal_id
  tenant_id    = azurerm_function_app.fapp.identity.tenant_id

  key_permissions    = ["Get"]
  secret_permissions = ["Get"]
}

# Storage Account
resource "azurerm_storage_account" "sa" {
  name                     = "sa${local.safe_group_name}"
  location                 = azurerm_resource_group.rg.location
  resource_group_name      = azurerm_resource_group.rg.name
  account_replication_type = "LRS"
  account_tier             = "Standard"
}

resource "azurerm_storage_queue" "sa" {
  name                 = "webhook-requests"
  storage_account_name = azurerm_storage_account.sa.name
}

resource "azurerm_storage_table" "sa" {
  name                 = "watch"
  storage_account_name = azurerm_storage_account.sa.name
}

resource "azurerm_storage_container" "sa" {
  name                 = "webhook-requests"
  storage_account_name = azurerm_storage_account.sa.name
}

resource "azurerm_key_vault_secret" "sa" {
  name         = "sa-connection-string"
  value        = azurerm_storage_account.sa.primary_connection_string
  key_vault_id = azurerm_key_vault.core.id
}

resource "azurerm_key_vault_secret" "sa" {
  name         = "sa-access-key"
  value        = azurerm_storage_account.sa.primary_access_key
  key_vault_id = azurerm_key_vault.core.id
}