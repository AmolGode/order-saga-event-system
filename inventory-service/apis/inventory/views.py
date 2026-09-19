from rest_framework.generics import ListAPIView

from .models import Inventory, ProcessedEvent, Product
from .serializers import InventorySerializer, ProcessedEventSerializer, ProductSerializer


class ProductListView(ListAPIView):
    queryset = Product.objects.all().order_by('name')
    serializer_class = ProductSerializer


class InventoryListView(ListAPIView):
    queryset = Inventory.objects.all().order_by('product_id')
    serializer_class = InventorySerializer


class ProcessedEventListView(ListAPIView):
    queryset = ProcessedEvent.objects.all().order_by('-processed_at')
    serializer_class = ProcessedEventSerializer
