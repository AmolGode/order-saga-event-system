from django.db import connection
from rest_framework import status
from rest_framework.generics import ListAPIView
from rest_framework.response import Response
from rest_framework.views import APIView

from .models import Payment, ProcessedEvent
from .serializers import PaymentSerializer, ProcessedEventSerializer


class HealthCheck(APIView):
    def get(self, request):
        try:
            with connection.cursor() as cursor:
                cursor.execute("SELECT 1")
        except Exception as exc:
            return Response({"status": "unhealthy", "error": str(exc)}, status=status.HTTP_503_SERVICE_UNAVAILABLE)
        return Response({"status": "ok"})


class PaymentListView(ListAPIView):
    queryset = Payment.objects.all().order_by('-created_at')
    serializer_class = PaymentSerializer


class ProcessedEventListView(ListAPIView):
    queryset = ProcessedEvent.objects.all().order_by('-processed_at')
    serializer_class = ProcessedEventSerializer
