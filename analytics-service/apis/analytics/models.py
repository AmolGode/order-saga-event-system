from django.db import models


class AnalyticsEvent(models.Model):
    event_id = models.UUIDField(primary_key=True)
    order_id = models.UUIDField()
    topic = models.CharField(max_length=100)
    status = models.CharField(max_length=20)
    timestamp = models.DateTimeField()
    payload = models.JSONField()
