
# Order REquest Post Api: Publish Event
curl -X POST http://localhost:8100/orders/create_order/ -H "Content-Type: application/json" -d '{"item":"test"}'




# Go Workers Required dependencies
go get github.com/confluentinc/confluent-kafka-go/v2/kafka@v2.15.1
go get github.com/jackc/pgx/v5/pgxpool@v5.11.0



# Curl to create Order 
curl -X POST http://localhost:8100/orders/create_order/ \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"11111111-1111-1111-1111-111111111111","product_id":"234bf452-a77c-48aa-86f4-bb0b358ba778","qty":2}'




