terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Local backend by default — fine for one person spinning this up and
  # tearing it down. Switch to an S3 backend (with DynamoDB state locking)
  # before more than one person applies against this, or before this becomes
  # a persistent environment rather than a throwaway load test.
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "order-saga-event-system"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

data "aws_availability_zones" "available" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
