# Matches your existing .env naming exactly — same DB/user names, same
# passwords by default, so nothing in the app config needs to change to
# point at RDS instead of the docker-compose Postgres containers.
locals {
  # instance_class pinned to db.t3.micro and allocated_storage capped at 20
  # on all 4 — Free Tier rejects both the previous larger instance classes
  # (confirmed: FreeTierRestrictionError) and anything over 20GB storage.
  databases = {
    order = {
      db_name           = "orders"
      username          = "order_user"
      password          = var.order_db_password
      instance_class    = "db.t3.micro"
      allocated_storage = 20
      iops              = 6000 # order-db takes the direct load-test traffic — provisioned IOPS, not gp3 baseline
    }
    inventory = {
      db_name           = "inventory"
      username          = "inventory_user"
      password          = var.inventory_db_password
      instance_class    = "db.t3.micro"
      allocated_storage = 20
      iops              = null
    }
    payment = {
      db_name           = "payments"
      username          = "payment_user"
      password          = var.payment_db_password
      instance_class    = "db.t3.micro"
      allocated_storage = 20
      iops              = null
    }
    analytics = {
      db_name           = "analytics"
      username          = "analytics_user"
      password          = var.analytics_db_password
      instance_class    = "db.t3.micro"
      allocated_storage = 20
      iops              = null
    }
  }
}

resource "aws_security_group" "rds" {
  name_prefix = "${var.cluster_name}-rds-"
  description = "Allow Postgres from the EKS cluster nodes and pods only."
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Postgres from EKS nodes"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "this" {
  name       = "${var.cluster_name}-db-subnets"
  subnet_ids = module.vpc.private_subnets
}

resource "aws_db_instance" "this" {
  for_each = local.databases

  identifier     = "${var.cluster_name}-${each.key}-db"
  engine         = "postgres"
  engine_version = "16"

  instance_class         = each.value.instance_class
  allocated_storage      = each.value.allocated_storage
  storage_type           = each.value.iops != null ? "io2" : "gp3"
  iops                   = each.value.iops
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  db_name  = each.value.db_name
  username = each.value.username
  password = each.value.password
  port     = 5432

  # Load-test throwaway env: skip the final snapshot so `terraform destroy`
  # doesn't hang waiting on one. Turn this back on before anything real.
  skip_final_snapshot = true
  publicly_accessible = false

  multi_az = false # single-AZ is fine for a load test; flip this on for a persistent env
}
