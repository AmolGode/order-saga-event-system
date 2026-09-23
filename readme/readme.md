order-saga-event-system/
├── .env
├── .gitignore
├── docker-compose.yml
├── README.md
│
├── order-service/
│   ├── api/                          (Django)
│   │   ├── manage.py
│   │   ├── requirements.txt
│   │   ├── Dockerfile
│   │   ├── config/
│   │   │   ├── settings.py
│   │   │   ├── urls.py
│   │   ├── orders/
│   │   │   ├── models.py
│   │   │   ├── views.py
│   │   │   ├── migrations/
│   │   ├── kafka/
│   │   │   ├── producer.py
│   │
│   └── worker/                       (Go)
│       ├── go.mod
│       ├── go.sum
│       ├── main.go
│       ├── consumer.go
│       ├── handlers.go
│       ├── db.go
│       ├── config.go
│       ├── Dockerfile
│
├── inventory-service/
│   ├── api/                          (Django — dashboard/manual stock CRUD)
│   │   ├── manage.py
│   │   ├── requirements.txt
│   │   ├── Dockerfile
│   │   ├── config/
│   │   ├── stock/
│   │   │   ├── models.py
│   │   │   ├── views.py
│   │   │   ├── migrations/
│   │
│   └── worker/                       (Go — reserve/commit/release stock)
│       ├── go.mod
│       ├── go.sum
│       ├── main.go
│       ├── consumer.go
│       ├── producer.go               (produces inventory-reserved/failed)
│       ├── handlers.go
│       ├── db.go
│       ├── config.go
│       ├── Dockerfile
│
├── payment-service/
│   └── worker/                       (Go — mock charge, produces payment-success/failed)
│       ├── go.mod
│       ├── go.sum
│       ├── main.go
│       ├── consumer.go
│       ├── producer.go
│       ├── handlers.go
│       ├── db.go
│       ├── config.go
│       ├── Dockerfile
│
├── notification-service/
│   ├── orchestrator/










**Order Saga Event System** — a microservices-based e-commerce order flow using the saga pattern, with Kafka as the event backbone. Repo: `order-saga-event-system` (monorepo).

**Pattern per service:** `api/` (Django, sync request-response/dashboard) + `worker/` (Go, async Kafka consumer/producer), each service owns its own DB.

| Service | api/ (Django) | worker/ (Go) |
|---|---|---|
| **order-service** | Create/view orders | Consume payment/inventory events → confirm or cancel the order |
| **inventory-service** | Stock CRUD | Consume `order.created` → reserve stock; consume cancel/fail events → release stock |
| **payment-service** | — | Consume `inventory.reserved` → mock-charge → publish `payment.succeeded`/`payment.failed` |
| **notification-service** | — | Orchestrator + multiple named workers (`email-worker`, `sms-worker`, `push-worker`, `whatsapp-worker`) → send notifications on saga events |
| **analytics-service** | — | Consume all events → write to Cassandra for reporting |

**Flow (saga):** order-service creates order → inventory-service reserves stock → payment-service charges → on success, order confirmed + customer notified; on failure at any step, upstream services get a cancel/release event to roll back (choreography-style saga, no central orchestrator except notification's own internal one).

**Infra:** Kafka (event bus), one root-level `docker-compose.yml`, Kubernetes later.









Final topic list with partition counts:

Topic	Partitions	Why
order-requested	10	Matches 10k/sec baseline — main pipeline
inventory-reserved	10	Same pipeline, same pace
inventory-failed	3	Failure-only, much lower volume
payment-success	10	Same main pipeline pace
payment-failed	3	Failure-only, lower volume
email-notifications	3	Bottlenecked by SES (~50/sec)
sms-notifications	2	Bottlenecked by Twilio (~1/sec/number), OTP-only
push-notifications	6	No real provider cap, higher volume
whatsapp-notifications	2	Bottlenecked by Meta (~80/sec), OTP-only











1. Client → Order svc API → creates Order (PENDING), generates event_id, 
   produces order-requested (key=order_id)

2. Inventory worker (1 of N pods) polls, gets the message
3. Check: SELECT * FROM ProcessedEvent WHERE event_id = X
   - exists → skip (already handled), commit offset, done
   - not exists → proceed
4. Reserve stock (UPDATE stock SET qty = qty - N) + INSERT ProcessedEvent(event_id)
   — both in ONE DB transaction
5. Commit Kafka offset









**Full flow, service by service:**

**1. Order Service (API)**
- Client sends `POST /order` with product(s) + quantity
- Creates `Order` row, status = PENDING
- Generates a unique `event_id`, publishes `order-requested` (key = `order_id`) to Kafka

**2. Inventory Service (worker)**
- Consumes `order-requested`
- Checks `ProcessedEvent` — skip if already handled (idempotency)
- Checks stock: enough → reserves it (atomic decrement), publishes `inventory-reserved`
- Not enough → publishes `inventory-failed`

**3. Payment Service (worker)** — only runs on success path
- Consumes `inventory-reserved`
- Checks `ProcessedEvent` — skip if duplicate
- Mock-charges the card
- Success → publishes `payment-success`
- Failure → publishes `payment-failed`

**4. Order Service (worker, same service as step 1 — different component)**
- Consumes `payment-success` → order status = CONFIRMED
- Consumes `payment-failed` OR `inventory-failed` → order status = CANCELLED

**5. Inventory Service (worker, additional subscription)** — the compensation step
- Also consumes `payment-failed`
- Releases the previously reserved stock back (undo step 2's reservation)

**This whole chain is the Saga pattern — specifically "choreography" style:** no central coordinator service exists; each service reacts to the event before it and emits the next one. The "some cases" (failure branches in steps 2-5) are the **compensating transactions** — the defining feature of a saga: when a later step fails, an earlier step's effect gets explicitly undone via its own event, rather than a database rollback (which isn't possible across separate DBs).




# Event IDS for each case

**payment-success (normal)** — order_id `4c08d298...`
- incoming payment-requested event_id: `057e0e2f-6160-5700-9c96-e480b36d8d9a`
- outcome event_id (payment-success): `3be7d629-f023-5cdc-8f7b-d48bca9b850d`
- order status: CONFIRMED

**inventory-failed (forced 0 stock)** — order_id `257d27c9...`
- incoming order-requested event_id: `0fb7daaa-9b9c-44e4-8194-290d8ede504d`
- outcome event_id (inventory-failed): `38dfb0db-ecbc-505e-929b-734190de8e09`
- order status: CANCELLED / "Insufficient inventory"

**payment-failed (forced decline)** — order_id `bba6037b...`
- incoming payment-requested event_id: `bc6c9eaf-9081-5f33-8537-241d9307f0df`
- outcome event_id (payment-failed): `8c9c0850-60ce-5313-a414-a1389f482c08`
- order status: CANCELLED / "Payment failed"

**payment-failed (DB down, technical failure)** — order_id `2f17872c...`
- incoming payment-requested event_id: `670e53a3-28e7-5c7f-a7b6-a290e8adf23f`
- outcome event_id (payment-failed): `7ddfd364-57cc-5dd2-b037-ac1e7728439c`
- order status: CANCELLED / "Payment failed" — raw message also preserved in `payment-requested-dlq`

Note: order-service itself doesn't read/store these event_ids (it only extracts order_id) — these come from analytics-service's independent tracking.

# Topics and partitions actually configured

(The "Final topic list with partition counts" table above is an old planning doc — it has topic names like `inventory-reserved` and notification-service topics that were never actually built. This section is the real, current state.)

**Local (docker-compose), currently on the broker:**
- order-requested — 2 partitions, replication factor 1
- inventory-failed — 2 partitions, replication factor 1
- payment-requested — 2 partitions, replication factor 1
- payment-success — 2 partitions, replication factor 1
- payment-failed — 2 partitions, replication factor 1
- order-requested-dlq — 3 partitions, replication factor 1
- payment-requested-dlq — 3 partitions, replication factor 1
- inventory-failed-dlq — 3 partitions, replication factor 1
- payment-success-dlq — 3 partitions, replication factor 1
- payment-failed-dlq — 3 partitions, replication factor 1

All 10 keyed by order_id (see partition-key fix above). Replication factor is 1 everywhere locally since it's a single-broker setup. `inventory-checked` was removed — matches the corrected HLD (`event-order-saga.excalidraw`): inventory-service publishes straight to `payment-requested` after reserving stock, no intermediate event.

**k8s (cloud/EKS), configured in `k8s/kafka/topics-job.yaml`:**
- Same 10 topics as above, all created with 3 partitions and replication factor 3 (matches the 3-broker StatefulSet, so a broker loss doesn't lose data — RF=1 locally can't do that since there's only 1 broker).









