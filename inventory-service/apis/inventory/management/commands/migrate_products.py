import json
from pathlib import Path
from django.core.management.base import BaseCommand
from inventory.models import Product, Inventory


class Command(BaseCommand):
    help = "Migrate test products and inventory data"

    def handle(self, *args, **options):
        # Fixed ids, not left to Product's default=uuid.uuid4 — these exact
        # values are hardcoded in order-service/apis/data/product_data.json
        # and in order_load_test/order_load_test.js's PRODUCT_IDS. If ids
        # were left random, every fresh database (a new RDS instance, a
        # teammate's local setup) would generate different ids than those
        # two files expect, and every order would 400 with "Unknown
        # product_id" the moment you pointed at a database this command
        # hadn't already seeded with those exact random values.
        #
        # Stock is 1,000,000,000 per product, not 100/50/200 — this seeds
        # for a sustained 10k/sec load test (see order_load_test.js), not a
        # quick manual check. At 10k/sec x ~3 avg qty/order, a 2-minute test
        # consumes ~3.6M units per product; 100 units would've been gone in
        # about 3 milliseconds, turning nearly every order into a false
        # inventory-failed instead of exercising the real saga path.
        products = [
            {"id": "234bf452-a77c-48aa-86f4-bb0b358ba778", "name": "Wireless Mouse", "sku": "SKU-001", "price": 599.00, "stock": 1_000_000_000},
            {"id": "d2350b96-e9fd-401d-8b47-3fa62f5249d9", "name": "Mechanical Keyboard", "sku": "SKU-002", "price": 2499.00, "stock": 1_000_000_000},
            {"id": "ee4c2554-84a3-4e78-94bc-8665326e5934", "name": "USB-C Cable", "sku": "SKU-003", "price": 199.00, "stock": 1_000_000_000},
        ]

        product_cache = []

        for p in products:
            product, _ = Product.objects.get_or_create(
                sku=p["sku"],
                defaults={"id": p["id"], "name": p["name"], "price": p["price"]},
            )
            Inventory.objects.get_or_create(
                product=product,
                defaults={"quantity_available": p["stock"], "quantity_reserved": 0},
            )
            self.stdout.write(self.style.SUCCESS(f"Seeded {product.name}"))

            product_cache.append({
                "id": str(product.id),
                "name": product.name,
                "sku": product.sku,
                "price": str(product.price),
            })

        output_path = Path(__file__).resolve().parent / "migrate_products_output.json"
        with open(output_path, "w") as f:
            json.dump(product_cache, f, indent=2)

        self.stdout.write(self.style.SUCCESS(f"Wrote product cache to {output_path}"))
