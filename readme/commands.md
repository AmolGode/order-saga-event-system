
# Order REquest Post Api: Publish Event
curl -X POST http://localhost:8100/orders/create_order/ -H "Content-Type: application/json" -d '{"item":"test"}'




# Go Workers Required dependencies
go get github.com/confluentinc/confluent-kafka-go/v2/kafka@v2.15.1
go get github.com/jackc/pgx/v5/pgxpool@v5.11.0



# Curl to create Order 
curl -X POST http://localhost:8100/orders/create_order/ \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"11111111-1111-1111-1111-111111111111","product_id":"234bf452-a77c-48aa-86f4-bb0b358ba778","qty":2}'




# See First Offset / Next Offset per partition (topic overview, via CLI)
docker compose exec kafka-broker /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server localhost:9092 --topic order-requested --time -2   # First Offset (earliest)
docker compose exec kafka-broker /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server localhost:9092 --topic order-requested --time -1   # Next Offset (latest)
# Message count = Next Offset - First Offset (per partition)