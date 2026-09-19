from rest_framework import serializers

from .models import Payment, ProcessedEvent


class PaymentSerializer(serializers.ModelSerializer):
    class Meta:
        model = Payment
        fields = '__all__'


class ProcessedEventSerializer(serializers.ModelSerializer):
    class Meta:
        model = ProcessedEvent
        fields = '__all__'
