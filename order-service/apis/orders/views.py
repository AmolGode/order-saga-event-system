import json
import uuid
from pathlib import Path
from django.conf import settings
from django.db import connection
from rest_framework.views import APIView
from rest_framework.generics import ListAPIView
from rest_framework.response import Response
from rest_framework import status
from kafka.producer import produce_event
from .models import Order, OrderItem, ProcessedEvent
from .serializers import OrderSerializer, OrderItemSerializer, ProcessedEventSerializer


class HealthCheck(APIView):
    """Checks the thing that actually breaks under load: can this pod reach
    Postgres right now? /metrics can't tell you that — it just dumps
    in-memory counters regardless of DB state."""

    def get(self, request):
        try:
            with connection.cursor() as cursor:
                cursor.execute("SELECT 1")
        except Exception as exc:
            return Response({"status": "unhealthy", "error": str(exc)}, status=status.HTTP_503_SERVICE_UNAVAILABLE)
        return Response({"status": "ok"})


class CreateOrder(APIView):
    def post(self, request):
        customer_id = request.data.get("customer_id")
        product_id = request.data.get("product_id")
        qty = request.data.get("qty")

        if not customer_id or not product_id or not qty:
            return Response(
                {"message": "customer_id, product_id and qty are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        product_data_path = Path(settings.BASE_DIR) / "data" / "product_data.json"
        with open(product_data_path) as f:
            products = {p["id"]: p for p in json.load(f)}

        product = products.get(product_id)
        if not product:
            return Response(
                {"message": f"Unknown product_id: {product_id}"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        unit_price = float(product["price"])
        total_amount = unit_price * qty

        order = Order.objects.create(customer_id=customer_id, total_amount=total_amount)

        OrderItem.objects.create(
            order=order,
            product_id=product_id,
            quantity=qty,
            unit_price=unit_price,
        )

        event_id = str(uuid.uuid4())

        produce_event("order-requested", {
            "event_id": event_id,
            "order_id": str(order.id),
            "customer_id": str(order.customer_id),
            "product_id": product_id,
            "qty": qty,
            "unit_price": unit_price,
            "total_amount": total_amount,
        },
        key=str(order.id)
        )

        return Response(
            {"message": "Order Request Added", "order_id": str(order.id), "event_id": event_id},
            status=status.HTTP_201_CREATED,
        )


class OrderListView(ListAPIView):
    queryset = Order.objects.all().order_by('-created_at')
    serializer_class = OrderSerializer


class OrderItemListView(ListAPIView):
    queryset = OrderItem.objects.all().order_by('-order_id')
    serializer_class = OrderItemSerializer


class ProcessedEventListView(ListAPIView):
    queryset = ProcessedEvent.objects.all().order_by('-processed_at')
    serializer_class = ProcessedEventSerializer
