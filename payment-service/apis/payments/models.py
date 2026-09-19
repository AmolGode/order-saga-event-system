from django.db import models


class Payment(models.Model):
    STATUS_CHOICES = [
        ('SUCCESS', 'Success'),
        ('FAILED', 'Failed'),
    ]

    order_id = models.UUIDField()
    amount = models.DecimalField(max_digits=10, decimal_places=2)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES)
    created_at = models.DateTimeField(auto_now_add=True)


class ProcessedEvent(models.Model):
    STATUS_CHOICES = [
        ('SUCCESS', 'Success'),
        ('FAILED', 'Failed'),
    ]

    event_id = models.UUIDField(primary_key=True)
    topic = models.CharField(max_length=100)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES, default='SUCCESS')
    processed_at = models.DateTimeField(auto_now_add=True)
    result = models.JSONField(null=True, blank=True)
