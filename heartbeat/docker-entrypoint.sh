#!/bin/bash
export ELASTICSEARCH_HOST=${ELASTICSEARCH_HOST:-http://elasticsearch:9200}
export KIBANA_HOST=${KIBANA_HOST:-http://kibana:5601}

heartbeat setup --index-management \
  -E setup.kibana.host=$KIBANA_HOST \
  -E output.elasticsearch.hosts=[$ELASTICSEARCH_HOST]


exec heartbeat -e \
  -E output.elasticsearch.hosts=[$ELASTICSEARCH_HOST] \
  -E setup.kibana.host=$KIBANA_HOST
