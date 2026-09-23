# k6-operator (distributed load test, in-cluster)

Runs the 10k/sec load test from *inside* the EKS cluster, across multiple k6 pods, instead
of from your laptop. Running it from a laptop against a real AWS endpoint means your home
network's bandwidth/latency becomes the bottleneck you're measuring, not the system —
this runs the generator in the same VPC as the thing it's testing.

## One-time setup: install the k6-operator

Same pattern as KEDA — the operator + its `TestRun` CRD is a separate, official install,
not something to hand-author here:

```bash
curl https://raw.githubusercontent.com/grafana/k6-operator/main/bundle.yaml | kubectl apply -f -
```

Wait for it to be ready:

```bash
kubectl get pods -n k6-operator-system -w
```

## What's in this folder

- `configmap.yaml` — the load test script (mirrors `order_load_test/order_load_test.js`
  exactly — k6-operator can only load a script from a ConfigMap/volume, not straight from
  this repo, so keep the two in sync by hand if the script changes).
- `testrun.yaml` — the `TestRun` that actually runs it: 5 runner pods, each targeting
  ~2k/sec (5 × 2k = the script's `rate: 10000` total, split automatically by k6-operator).

**Scheduling:** `runner.nodeSelector: role: app` pins the load-generator pods to the app
node group. No toleration is added for the Kafka node group's `dedicated=kafka:NoSchedule`
taint — that taint exists specifically to keep non-broker pods off those nodes, so a k6
pod without a matching toleration simply can't schedule there.

**Cost note:** these are *additional* pods on top of your app/worker pods, so they need
their own node capacity for the test's duration (5 pods × ~1-2 CPU each ≈ 5-10 extra vCPU
worth of headroom on the app node group's Spot instances). Likely still a trivial add-on
to your existing load-test budget, but worth remembering when you check the AWS Billing
dashboard after the run — it's easy to forget the load generator itself costs something
too, not just the system under test.

## Running it

```bash
kubectl apply -f k8s/k6-operator/configmap.yaml
kubectl apply -f k8s/k6-operator/testrun.yaml
```

Watch progress:

```bash
kubectl get testrun order-saga-load-test -w
kubectl logs -l k6_cr=order-saga-load-test,runner=true -f --prefix
```

Clean up after the run (a `TestRun` doesn't auto-delete its pods):

```bash
kubectl delete -f k8s/k6-operator/testrun.yaml
```
