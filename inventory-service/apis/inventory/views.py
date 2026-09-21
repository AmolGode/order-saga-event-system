from django.db import connection
from rest_framework import status
from rest_framework.generics import ListAPIView
from rest_framework.response import Response
from rest_framework.views import APIView

from .models import Inventory, ProcessedEvent, Product
from .serializers import InventorySerializer, ProcessedEventSerializer, ProductSerializer


class HealthCheck(APIView):
    def get(self, request):
        try:
            with connection.cursor() as cursor:
                cursor.execute("SELECT 1")
        except Exception as exc:
            return Response({"status": "unhealthy", "error": str(exc)}, status=status.HTTP_503_SERVICE_UNAVAILABLE)
        return Response({"status": "ok"})


class ProductListView(ListAPIView):
    queryset = Product.objects.all().order_by('name')
    serializer_class = ProductSerializer


class InventoryListView(ListAPIView):
    queryset = Inventory.objects.all().order_by('product_id')
    serializer_class = InventorySerializer


class ProcessedEventListView(ListAPIView):
    queryset = ProcessedEvent.objects.all().order_by('-processed_at')
    serializer_class = ProcessedEventSerializer
