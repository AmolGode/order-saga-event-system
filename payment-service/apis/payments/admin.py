from django.contrib import admin
from .models import Payment, ProcessedEvent


@admin.register(Payment)
class PaymentAdmin(admin.ModelAdmin):
    list_display = ("id", "order_id", "amount", "status", "created_at")


@admin.register(ProcessedEvent)
class ProcessedEventAdmin(admin.ModelAdmin):
    list_display = ("event_id", "topic", "status", "processed_at")
