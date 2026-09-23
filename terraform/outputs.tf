output "cluster_name" {
  value = module.eks.cluster_name
}

output "configure_kubectl" {
  description = "Run this after apply to point kubectl at the new cluster."
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "db_endpoints" {
  description = "Direct RDS endpoints for all 4 databases — use these as POSTGRES_HOST. No RDS Proxy in front of any of them; PgBouncer (see docker-compose.yml) already handles order-db's connection pooling."
  value       = { for k, v in aws_db_instance.this : k => v.endpoint }
}

output "kafka_node_group_taint" {
  description = "Reminder: Kafka broker pods need this toleration to actually schedule onto the kafka node group."
  value       = "tolerations: [{ key: dedicated, operator: Equal, value: kafka, effect: NoSchedule }]"
}
