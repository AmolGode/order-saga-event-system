# Monitoring (Prometheus + Grafana, in-cluster)

Ports Prometheus/Grafana from local docker-compose to the real cluster — same datasource,
same dashboard, but scrapes pods via Kubernetes service discovery instead of a static
target list, since KEDA (`k8s/keda/`) scales workers to a variable number of replicas.
A static target would only ever hit one pod behind each Service and silently miss
whatever else KEDA spun up.

## How discovery works

Every api/worker Deployment across the 4 services now carries
`prometheus.io/scrape: "true"` and `prometheus.io/port: "<port>"` annotations on its pod
template. `prometheus-configmap.yaml`'s single `kubernetes-pods` job discovers any pod
with that annotation, at whatever replica count exists right now, and labels each one
`job=<app-label>` — so `job="order-service-api"` etc. still work exactly like the local
static config did; the dashboard JSON needed zero changes.

`prometheus-rbac.yaml` grants the ServiceAccount the list/watch permissions this
discovery needs — that's the one piece with no local equivalent, since docker-compose
never needed an API server to ask "what pods exist right now."

## Apply order

Deploy after the 4 services and Kafka are already up, so there's something to discover:

```bash
kubectl apply -f k8s/monitoring/prometheus-rbac.yaml
kubectl apply -f k8s/monitoring/prometheus-configmap.yaml -f k8s/monitoring/prometheus-pvc.yaml -f k8s/monitoring/prometheus-deployment.yaml -f k8s/monitoring/prometheus-service.yaml
kubectl apply -f k8s/monitoring/kafka-exporter-deployment.yaml -f k8s/monitoring/kafka-exporter-service.yaml
kubectl apply -f k8s/monitoring/grafana-secret.yaml -f k8s/monitoring/grafana-configmap.yaml -f k8s/monitoring/grafana-pvc.yaml -f k8s/monitoring/grafana-deployment.yaml -f k8s/monitoring/grafana-service.yaml
```

Grafana: NodePort 30000 (mirrors local's 3000), login `admin`/`admin` — same plaintext
Secret caveat as everywhere else in this repo, fine for learning, not for real use.
Prometheus: NodePort 30090 (mirrors local's 9090).

## Keeping in sync

`prometheus-configmap.yaml` and `grafana-configmap.yaml` are hand-copied from
`monitoring/` at the repo root — if you change the local `prometheus.yml` scrape configs
or the Grafana dashboard JSON, update both places.
