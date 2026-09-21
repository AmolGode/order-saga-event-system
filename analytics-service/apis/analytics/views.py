from django.db import connection
from rest_framework import status
from rest_framework.views import APIView
from rest_framework.response import Response

from .models import AnalyticsEvent


class HealthCheck(APIView):
    def get(self, request):
        try:
            with connection.cursor() as cursor:
                cursor.execute("SELECT 1")
        except Exception as exc:
            return Response({"status": "unhealthy", "error": str(exc)}, status=status.HTTP_503_SERVICE_UNAVAILABLE)
        return Response({"status": "ok"})


def serialize_event(e):
    return {
        'event_id': str(e.event_id),
        'order_id': str(e.order_id),
        'topic': e.topic,
        'status': e.status,
        'timestamp': e.timestamp.isoformat(),
        'payload': e.payload,
    }


def derive_order_status(last_topic):
    if last_topic == 'payment-success':
        return 'CONFIRMED'
    if last_topic in ('payment-failed', 'inventory-failed'):
        return 'CANCELLED'
    return 'PENDING'


class EventListView(APIView):
    """GET /api/events/ — recent events, or ?order_id=<id> / ?event_id=<id> for one order's full journey.

    event_id is unique per hop (a fresh one is generated at every republish), so
    searching by it resolves to that hop's order_id first, then returns every
    event for that order — the actual journey, not just the one matching row.
    """

    def get(self, request):
        order_id = request.query_params.get('order_id')
        event_id = request.query_params.get('event_id')

        if not order_id and event_id:
            hop = AnalyticsEvent.objects.filter(event_id=event_id).first()
            order_id = str(hop.order_id) if hop else None
            if not order_id:
                return Response([])

        if order_id:
            qs = AnalyticsEvent.objects.filter(order_id=order_id).order_by('timestamp')
        else:
            qs = AnalyticsEvent.objects.exclude(topic__endswith='-dlq').order_by('-timestamp')[:200]

        return Response([serialize_event(e) for e in qs])


class OrderListView(APIView):
    """GET /api/orders/ — one row per order_id, status derived from its latest event."""

    def get(self, request):
        events = AnalyticsEvent.objects.exclude(topic__endswith='-dlq').order_by('order_id', '-timestamp')

        latest_by_order = {}
        for e in events:
            oid = str(e.order_id)
            if oid not in latest_by_order:
                latest_by_order[oid] = e

        rows = sorted(latest_by_order.values(), key=lambda e: e.timestamp, reverse=True)
        return Response([
            {
                'order_id': str(e.order_id),
                'status': derive_order_status(e.topic),
                'last_topic': e.topic,
                'last_event_at': e.timestamp.isoformat(),
            }
            for e in rows
        ])


class DeadLetterListView(APIView):
    """GET /api/dead-letters/ — events that exhausted retries and landed on a *-dlq topic."""

    def get(self, request):
        qs = AnalyticsEvent.objects.filter(topic__endswith='-dlq').order_by('-timestamp')[:200]
        return Response([serialize_event(e) for e in qs])
