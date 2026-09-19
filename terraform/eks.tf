module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.31"

  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # Load-test convenience — you'll be applying/curling from your own machine.
  # Tighten this to your IP range before this is anything but a throwaway env.
  cluster_endpoint_public_access = true

  eks_managed_node_groups = {
    # Django/Go app pods — stateless, cheap to interrupt, so Spot makes sense.
    app = {
      instance_types = var.app_instance_types
      capacity_type  = "SPOT"

      min_size     = var.app_node_min_size
      max_size     = var.app_node_max_size
      desired_size = var.app_node_desired_size

      labels = {
        role = "app"
      }
    }

    # Kafka brokers — hold state on local EBS volumes; a Spot interruption
    # mid-test loses a broker and can corrupt the run. On-Demand, tainted so
    # only broker pods (with a matching toleration) land here.
    kafka = {
      instance_types = [var.kafka_instance_type]
      capacity_type  = "ON_DEMAND"

      min_size     = var.kafka_broker_count
      max_size     = var.kafka_broker_count
      desired_size = var.kafka_broker_count

      labels = {
        role = "kafka"
      }

      taints = {
        kafka = {
          key    = "dedicated"
          value  = "kafka"
          effect = "NO_SCHEDULE"
        }
      }
    }
  }
}
