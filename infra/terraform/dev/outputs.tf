output "resource_group" {
  value = azurerm_resource_group.this.name
}

output "openai_account" {
  value = azurerm_cognitive_account.openai.name
}

output "openai_endpoint" {
  value = azurerm_cognitive_account.openai.endpoint
}

output "embed_deployment" {
  value = azurerm_cognitive_deployment.embed.name
}

output "chat_deployment" {
  value = azurerm_cognitive_deployment.chat.name
}

output "key_command" {
  description = "Prints the API key without storing it in state output."
  value       = "az cognitiveservices account keys list -g ${azurerm_resource_group.this.name} -n ${azurerm_cognitive_account.openai.name} --query key1 -o tsv"
}

output "app_url" {
  value = "https://${azurerm_container_app.this.ingress[0].fqdn}"
}
