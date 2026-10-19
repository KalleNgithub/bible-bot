# Terraform state bootstrap

Creates a general-purpose Azure storage account for Terraform state. Run once, before
`infra/terraform/dev`. Other projects can reuse it with their own state `key`.

This configuration uses local state on purpose: it manages one storage account, so
losing the state means re-importing one resource.

```bash
az login
terraform init
terraform plan -var subscription_id=<id>
# apply only after reviewing the plan
terraform output backend_config
```

Put the printed backend block into `infra/terraform/dev/versions.tf`.
