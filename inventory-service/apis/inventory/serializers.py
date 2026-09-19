from rest_framework import serializers

from .models import Inventory, ProcessedEvent, Product


class ProductSerializer(serializers.ModelSerializer):
    class Meta:
        model = Product
        fields = '__all__'


class InventorySerializer(serializers.ModelSerializer):
    class Meta:
        model = Inventory
        fields = '__all__'


class ProcessedEventSerializer(serializers.ModelSerializer):
    class Meta:
        model = ProcessedEvent
        fields = '__all__'
