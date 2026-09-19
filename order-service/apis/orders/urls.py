from django.urls import path
from .views import CreateOrder

urlpatterns = [
    path('create_order/', CreateOrder.as_view()),
]