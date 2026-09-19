import json
import random
import uuid
from pathlib import Path
from django.conf import settings
from django.core.management.base import BaseCommand
from django.test import Client

# docker compose exec order-service-api python manage.py test_create_order

class Command(BaseCommand):
    help = "Fire a test order against create_order using a random product"

    def handle(self, *args, **options):
        product_data_path = Path(settings.BASE_DIR) / "data" / "product_data.json"
        with open(product_data_path) as f:
            products = json.load(f)

        product = random.choice(products)
        payload = {
            "customer_id": str(uuid.uuid4()),
            "product_id": product["id"],
            "qty": random.randint(1, 5),
        }

        client = Client()
        response = client.post(
            "/orders/create_order/",
            data=json.dumps(payload),
            content_type="application/json",
        )

        self.stdout.write(f"Status: {response.status_code}")
        self.stdout.write(response.content.decode())



        

