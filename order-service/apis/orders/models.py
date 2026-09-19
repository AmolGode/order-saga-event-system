from django.db import models
import uuid



class Order(models.Model):
    STATUS_CHOICES = [
        ('PENDING', 'Pending'),
        ('CONFIRMED', 'Confirmed'),
        ('CANCELLED', 'Cancelled'),
    ]

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    customer_id = models.UUIDField()
    status = models.CharField(max_length=20, choices=STATUS_CHOICES, default='PENDING')
    total_amount = models.DecimalField(max_digits=10, decimal_places=2)
    failed_reason = models.CharField(max_length=255, blank=True, default='')
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)


class OrderItem(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    order = models.ForeignKey(Order, related_name='items', on_delete=models.CASCADE)
    product_id = models.UUIDField()
    quantity = models.PositiveIntegerField()
    unit_price = models.DecimalField(max_digits=10, decimal_places=2)





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