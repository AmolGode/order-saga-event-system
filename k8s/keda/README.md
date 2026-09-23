# KEDA (Kafka-lag-based autoscaling)

Scales `order-service-worker`, `inventory-service-worker`, and `payment-service-worker`
based on actual Kafka consumer group lag, instead of a fixed replica count or CPU usage.
Under real load, a worker falling behind on its topic is a much more direct signal to
scale on than CPU — CPU can look fine while a worker is still stuck waiting on Postgres,
for example.

## One-time setup: install the KEDA operator

KEDA itself (the operator + its CRDs — `ScaledObject`, `TriggerAuthentication`, etc.) is
a separate install from anything in this repo, and it ships too many CRDs to hand-author
here. Install the official release once per cluster, before applying anything in this
folder:

```bash
kubectl apply --server-side -f https://github.com/kedacore/keda/releases/download/v2.16.0/keda-2.16.0.yaml
```

Wait for it to be ready:

```bash
kubectl get pods -n keda -w
```

## What's in this folder

- `order-service-worker-scaledobject.yaml`
- `inventory-service-worker-scaledobject.yaml`
- `payment-service-worker-scaledobject.yaml`

Each one targets its matching Deployment (already defined in the sibling
`order-service/`, `inventory-service/`, `payment-service/` folders) and scales it based
on the total consumer-group lag across every topic that worker subscribes to.

`topic` is intentionally left unset on the Kafka trigger — the KEDA Kafka scaler then
sums lag across *all* topics the consumer group is actually subscribed to, which matters
here since e.g. `inventory-service-worker` consumes both `order-requested` (10 partitions)
and `payment-failed` (3 partitions) under one consumer group.

`maxReplicaCount: 10` matches the highest partition count among each worker's subscribed
topics (see `k8s/kafka/topics-job.yaml`) — going higher wouldn't help, since a partition
can only be assigned to one consumer at a time within a group.

`minReplicaCount: 2` (not 0) — scaling to zero would mean a burst of traffic sits
completely unprocessed until a cold-start pod comes up, which defeats the point of a
sustained 10k/sec test.

## Applying

```bash
kubectl apply -f k8s/keda/order-service-worker-scaledobject.yaml
kubectl apply -f k8s/keda/inventory-service-worker-scaledobject.yaml
kubectl apply -f k8s/keda/payment-service-worker-scaledobject.yaml
```

Check status:

```bash
kubectl get scaledobject
kubectl get hpa   # KEDA creates a standard HPA under the hood for each ScaledObject
```
