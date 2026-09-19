import uuid
from django.db import models


class Product(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    name = models.CharField(max_length=255)
    sku = models.CharField(max_length=50, unique=True)
    price = models.DecimalField(max_digits=10, decimal_places=2)
    created_at = models.DateTimeField(auto_now_add=True)


class Inventory(models.Model):
    product = models.OneToOneField(Product, on_delete=models.CASCADE, related_name='inventory')
    quantity_available = models.PositiveIntegerField(default=0)
    quantity_reserved = models.PositiveIntegerField(default=0)
    updated_at = models.DateTimeField(auto_now=True)



class ProcessedEvent(models.Model):
    STATUS_CHOICES = [
            ('SUCCESS', 'Success'),
            ('FAILED', 'Failed')
        ]
    
    event_id = models.UUIDField(primary_key=True)
    topic = models.CharField(max_length=100)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES,default='SUCCESS')
    processed_at = models.DateTimeField(auto_now_add=True)
    result = models.JSONField(null=True, blank=True)