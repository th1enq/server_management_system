#!/bin/bash

echo "Setting up Kibana index pattern for VCS SMS logs..."

# Wait for Kibana to be ready
echo "Waiting for Kibana to be ready..."
until curl -s http://localhost:5601/api/status > /dev/null; do
    echo "Waiting for Kibana..."
    sleep 5
done

echo "Kibana is ready!"

# Create index pattern
echo "Creating index pattern for vcs-sms-server-checks-*..."

curl -X POST "localhost:5601/api/saved_objects/index-pattern/vcs-sms-server-checks-pattern" \
-H "Content-Type: application/json" \
-H "kbn-xsrf: true" \
-d '{
  "attributes": {
    "title": "vcs-sms-server-checks-*",
    "timeFieldName": "@timestamp"
  }
}'

echo ""
echo "Index pattern created!"
echo "You can now view your server check logs in Kibana at http://localhost:5601"
echo "Index pattern: vcs-sms-server-checks-*"
