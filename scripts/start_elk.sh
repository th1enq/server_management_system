#!/bin/bash

echo "Starting ELK Stack for VCS SMS..."

# Start the services
docker-compose up -d elasticsearch
echo "Waiting for Elasticsearch to be ready..."
sleep 15

docker-compose up -d logstash
echo "Waiting for Logstash to be ready..."
sleep 15

docker-compose up -d kibana
echo "Waiting for Kibana to be ready..."
sleep 15

echo "ELK Stack is starting up!"
echo ""
echo "Services:"
echo "- Elasticsearch: http://localhost:9200"
echo "- Logstash: http://localhost:9600"
echo "- Kibana: http://localhost:5601"
echo ""
echo "Check service status with: docker-compose ps"
echo "View logs with: docker-compose logs -f [service-name]"
