variable "region" {
  description = "AWS region for the load-test environment."
  type        = string
  default     = "ap-south-1"
}

variable "cluster_name" {
  description = "EKS cluster name."
  type        = string
  default     = "order-saga-loadtest"
}

variable "cluster_version" {
  description = "Kubernetes version for EKS."
  type        = string
  default     = "1.30"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.42.0.0/16"
}

variable "environment" {
  description = "Tag applied to every resource — makes teardown-by-tag straightforward."
  type        = string
  default     = "loadtest"
}

# --- App node group (Spot) ---

variable "app_instance_types" {
  description = "Instance types for the Spot-backed app node group (Django/Go pods)."
  type        = list(string)
  default     = ["m6i.xlarge", "m6a.xlarge", "m5.xlarge"]
}

variable "app_node_min_size" {
  type    = number
  default = 3
}

variable "app_node_max_size" {
  description = "Ceiling for the app node group. Sized generously since this is a short-lived test, not a persistent budget."
  type        = number
  default     = 20
}

variable "app_node_desired_size" {
  type    = number
  default = 4
}

# --- Kafka node group (On-Demand — brokers hold state, Spot interruption mid-test is not acceptable) ---

variable "kafka_instance_type" {
  description = "Instance type for Kafka broker nodes — memory + disk throughput matter more than CPU here."
  type        = string
  default     = "r6i.xlarge"
}

variable "kafka_broker_count" {
  description = "Number of Kafka broker pods/nodes. 3 is the minimum for a real quorum."
  type        = number
  default     = 3
}

# --- RDS ---

variable "order_db_instance_class" {
  description = "order-db sees the direct load-test traffic, so it's sized up separately from the other three DBs."
  type        = string
  default     = "db.r6g.xlarge"
}

variable "secondary_db_instance_class" {
  description = "inventory-db/payment-db/analytics-db only see Kafka-driven traffic, not the direct HTTP load — stay small."
  type        = string
  default     = "db.t4g.medium"
}

# --- DB passwords ---
# No default on purpose — a plaintext default here would be a real secret
# checked into version control regardless of `sensitive = true` (that only
# redacts CLI/plan output). Pass these via a gitignored terraform.tfvars
# (copy terraform.tfvars.example) or pull from Secrets Manager/SSM instead.

variable "order_db_password" {
  type      = string
  sensitive = true
}

variable "inventory_db_password" {
  type      = string
  sensitive = true
}

variable "payment_db_password" {
  type      = string
  sensitive = true
}

variable "analytics_db_password" {
  type      = string
  sensitive = true
}
