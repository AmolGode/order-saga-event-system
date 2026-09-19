# kafka docker hub image : 4.3.1
https://hub.docker.com/r/apache/kafka?tag=4.3.1




# Kafka Confluent - Python client to work with kafka 
https://docs.confluent.io/kafka-clients/python/current/overview.html#installation


# Install Client
- cd order-service/apis
- source venv/bin/activate
- pip install confluent-kafka

- pip freeze > requirements.txt

- docker compose build order-service-api


# Kafka UI
- https://hub.docker.com/r/kafbat/kafka-ui