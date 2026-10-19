variable "subscription_id" {
  description = "Azure subscription to deploy into."
  type        = string
}

variable "location" {
  description = "Azure region. Must offer both OpenAI models as GlobalStandard."
  type        = string
  default     = "westeurope"
}

variable "prefix" {
  description = "Name prefix for all resources."
  type        = string
  default     = "bible"
}

variable "environment" {
  description = "Environment name used in resource names."
  type        = string
  default     = "dev"
}

variable "monthly_budget_eur" {
  description = "Monthly cost alert threshold. Alerts only; Azure does not hard-stop spend."
  type        = number
  default     = 5
}

variable "budget_alert_email" {
  description = "Address that receives budget alerts."
  type        = string
}

variable "embed_capacity" {
  description = "Embedding deployment capacity in thousands of tokens per minute."
  type        = number
  default     = 120
}

variable "chat_model" {
  description = "Chat model for query translation and language detection."
  type        = string
  default     = "gpt-5.4-nano"
}

variable "chat_model_version" {
  description = "Version of the chat model."
  type        = string
  default     = "2026-03-17"
}

variable "chat_capacity" {
  description = "Chat deployment capacity in thousands of tokens per minute."
  type        = number
  default     = 10
}

variable "image" {
  description = "Public bible-bot image repository."
  type        = string
  default     = "ghcr.io/kallengithub/bible-bot"
}

variable "image_tag" {
  description = "Tag of the bible-bot image to run. Set in terraform.tfvars."
  type        = string
}

variable "max_replicas" {
  description = "Upper bound on replicas, which also bounds cost under load."
  type        = number
  default     = 2
}
