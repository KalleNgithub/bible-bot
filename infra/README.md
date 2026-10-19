# Infrastructure

Terraform for the Azure resources. Apply is always run manually by the owner.

- `terraform/bootstrap`: shared state storage account (local state, run once).
- `terraform/dev`: resource group, Azure OpenAI account with embedding and chat deployments, monthly budget alert, Container Apps environment and the app (min replicas 0). The app image is pulled from public GHCR. State key `bible-dev.tfstate`.

```bash
cd terraform/dev
cp terraform.tfvars.example terraform.tfvars   # set subscription_id and budget_alert_email
terraform init
terraform plan
# apply only after reviewing the plan
```

Read the API key without storing it:

```bash
export AZURE_OPENAI_KEY=$(terraform output -raw key_command | sh)
```

## Deploy the app

The image is public on GHCR (only code and public-domain Bible data, no secrets). Every deploy, including the first, from the repo root:

```bash
tag=$(date +%Y%m%d%H%M) &&
docker build -t ghcr.io/kallengithub/bible-bot:$tag . &&
docker push ghcr.io/kallengithub/bible-bot:$tag &&
sed -i "s/^image_tag .*/image_tag          = \"$tag\"/" infra/terraform/dev/terraform.tfvars &&
(cd infra/terraform/dev && terraform apply && curl "$(terraform output -raw app_url)/?q=John%2011:25")
```

One-time setup before the first deploy:

```bash
gh auth refresh -s write:packages
gh auth token | docker login ghcr.io -u KalleNgithub --password-stdin
az provider register --namespace Microsoft.App --wait
```

After the first push, set the package visibility to public in its GitHub package settings; a public repo does not make the image public. If a failed apply leaves `ca-bible-dev` in state `Failed`, delete it with `az containerapp delete -n ca-bible-dev -g rg-bible-dev --yes` and apply again.

To roll back, set an earlier tag in `terraform.tfvars` and apply.
