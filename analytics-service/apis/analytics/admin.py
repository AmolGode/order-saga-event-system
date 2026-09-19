from django.contrib import admin
from .models import AnalyticsEvent


@admin.register(AnalyticsEvent)
class AnalyticsEventAdmin(admin.ModelAdmin):
    list_display = ("event_id", "order_id", "topic", "status", "timestamp")
