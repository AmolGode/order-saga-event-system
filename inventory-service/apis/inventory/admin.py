from django.contrib import admin
from .models import Product, Inventory, ProcessedEvent


@admin.register(Product)
class ProductAdmin(admin.ModelAdmin):
    list_display = ("id", "name", "sku", "price", "created_at")


@admin.register(Inventory)
class InventoryAdmin(admin.ModelAdmin):
    list_display = ("product", "quantity_available", "quantity_reserved", "updated_at")


@admin.register(ProcessedEvent)
class ProcessedEventAdmin(admin.ModelAdmin):
    list_display = ("event_id", "topic", "status", "processed_at")
