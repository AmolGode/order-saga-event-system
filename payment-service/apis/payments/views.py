from rest_framework.generics import ListAPIView

from .models import Payment, ProcessedEvent
from .serializers import PaymentSerializer, ProcessedEventSerializer


class PaymentListView(ListAPIView):
    queryset = Payment.objects.all().order_by('-created_at')
    serializer_class = PaymentSerializer


class ProcessedEventListView(ListAPIView):
    queryset = ProcessedEvent.objects.all().order_by('-processed_at')
    serializer_class = ProcessedEventSerializer
