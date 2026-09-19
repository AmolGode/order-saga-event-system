import json
from pathlib import Path
from django.core.management.base import BaseCommand
from inventory.models import Product, Inventory


class Command(BaseCommand):
    help = "Migrate test products and inventory data"

    def handle(self, *args, **options):
        products = [
            {"name": "Wireless Mouse", "sku": "SKU-001", "price": 599.00, "stock": 100},
            {"name": "Mechanical Keyboard", "sku": "SKU-002", "price": 2499.00, "stock": 50},
            {"name": "USB-C Cable", "sku": "SKU-003", "price": 199.00, "stock": 200},
        ]

        product_cache = []

        for p in products:
            product, _ = Product.objects.get_or_create(
                sku=p["sku"],
                defaults={"name": p["name"], "price": p["price"]},
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
