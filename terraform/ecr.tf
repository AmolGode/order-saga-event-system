# One repo per image each service builds. notification-service has no repo
# since it doesn't exist in this codebase.
locals {
  ecr_repos = [
    "order-service-api", "order-service-worker",
    "inventory-service-api", "inventory-service-worker",
    "payment-service-api", "payment-service-worker",
    "analytics-service-api", "analytics-service-worker",
  ]
}

resource "aws_ecr_repository" "this" {
  for_each = toset(local.ecr_repos)

  name                 = each.value
  image_tag_mutability = "IMMUTABLE" # forces a real tag per build instead of overwriting :latest

  image_scanning_configuration {
    scan_on_push = true
  }
}

output "ecr_repository_urls" {
  description = "Push images here, then reference these URLs (with your build's tag) in the k8s Deployment manifests."
  value       = { for k, v in aws_ecr_repository.this : k => v.repository_url }
}
