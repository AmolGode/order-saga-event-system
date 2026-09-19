# One repo per image order-service actually builds. Extend this map when the
# other services' k8s manifests get built out.
locals {
  ecr_repos = ["order-service-api", "order-service-worker"]
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
