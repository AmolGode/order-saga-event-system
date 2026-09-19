output "cluster_name" {
  value = module.eks.cluster_name
}

output "configure_kubectl" {
  description = "Run this after apply to point kubectl at the new cluster."
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "order_db_proxy_endpoint" {
  description = "POSTGRES_HOST for order-service-api/worker to use instead of the raw RDS endpoint."
  value       = aws_db_proxy.order_db.endpoint
}

output "db_endpoints" {
  description = "Direct RDS endpoints — used for inventory/payment/analytics, which don't go through a proxy."
  value       = { for k, v in aws_db_instance.this : k => v.endpoint }
}

output "kafka_node_group_taint" {
  description = "Reminder: Kafka broker pods need this toleration to actually schedule onto the kafka node group."
  value       = "tolerations: [{ key: dedicated, operator: Equal, value: kafka, effect: NoSchedule }]"
}
