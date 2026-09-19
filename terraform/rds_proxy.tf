# RDS Proxy replaces PgBouncer for the AWS version — it's the managed,
# HA equivalent, and only order-db needs it: that's the one instance facing
# 50-100+ app pods each wanting a connection, same reason PgBouncer went in
# front of order-db locally and nowhere else.

resource "aws_secretsmanager_secret" "order_db" {
  name_prefix = "${var.cluster_name}-order-db-"
}

resource "aws_secretsmanager_secret_version" "order_db" {
  secret_id = aws_secretsmanager_secret.order_db.id
  secret_string = jsonencode({
    username = local.databases.order.username
    password = local.databases.order.password
  })
}

resource "aws_iam_role" "rds_proxy" {
  name_prefix = "${var.cluster_name}-rds-proxy-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "rds.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "rds_proxy_secrets" {
  name = "read-order-db-secret"
  role = aws_iam_role.rds_proxy.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["secretsmanager:GetSecretValue"]
      Resource = [aws_secretsmanager_secret.order_db.arn]
    }]
  })
}

resource "aws_db_proxy" "order_db" {
  name                   = "${var.cluster_name}-order-db-proxy"
  engine_family          = "POSTGRESQL"
  role_arn               = aws_iam_role.rds_proxy.arn
  vpc_subnet_ids         = module.vpc.private_subnets
  vpc_security_group_ids = [aws_security_group.rds.id]

  auth {
    auth_scheme = "SECRETS"
    secret_arn  = aws_secretsmanager_secret.order_db.arn
    iam_auth    = "DISABLED"
  }
}

resource "aws_db_proxy_default_target_group" "order_db" {
  db_proxy_name = aws_db_proxy.order_db.name

  connection_pool_config {
    # Real ceiling on concurrent Postgres connections regardless of how
    # many app pods are hammering the proxy — same role DEFAULT_POOL_SIZE
    # played for PgBouncer locally.
    max_connections_percent = 90
  }
}

resource "aws_db_proxy_target" "order_db" {
  db_proxy_name          = aws_db_proxy.order_db.name
  target_group_name      = aws_db_proxy_default_target_group.order_db.name
  db_instance_identifier = aws_db_instance.this["order"].identifier
}
