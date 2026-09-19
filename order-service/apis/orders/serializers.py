from rest_framework import serializers

from .models import Order, OrderItem, ProcessedEvent


class OrderSerializer(serializers.ModelSerializer):
    class Meta:
        model = Order
        fields = '__all__'


class OrderItemSerializer(serializers.ModelSerializer):
    class Meta:
        model = OrderItem
        fields = '__all__'


class ProcessedEventSerializer(serializers.ModelSerializer):
    class Meta:
        model = ProcessedEvent
        fields = '__all__'
