#!/bin/bash

# Start monitoring stack (Loki + Grafana + Prometheus + Promtail)
echo "Starting monitoring stack..."

cd docker
docker-compose -f monitoring-compose.yml up -d

echo "Waiting for services to be ready..."
sleep 15

echo "Monitoring stack is ready!"
echo ""
echo "Access URLs:"
echo "  Grafana:    http://localhost:3000 (admin/admin)"
echo "  Prometheus: http://localhost:9090"
echo "  Loki:       http://localhost:3100"
echo ""
echo "To stop the monitoring stack, run: docker-compose -f monitoring-compose.yml down"
