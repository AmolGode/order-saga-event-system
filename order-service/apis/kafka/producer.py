from confluent_kafka import Producer
from django.conf import settings
import socket
import json

_producer = None

def get_producer():
    global _producer
    if _producer is None:
        conf = {
            'bootstrap.servers':settings.KAFKA_BOOTSTRAP_SERVERS,
            'client.id': socket.gethostname() # will help in.kafka's logs/metrics
        }
        _producer = Producer(conf)
    return _producer


def produce_event(topic:str, payload:dict, key:str):
    producer = get_producer()
    producer.produce(topic, key=key, value=json.dumps(payload).encode("utf-8"))
