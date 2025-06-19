ARG ELK_VERSION=9.0.2
FROM docker.elastic.co/beats/heartbeat:${ELK_VERSION}

USER root

COPY config/heartbeat.yml /usr/share/heartbeat/heartbeat.yml

RUN chmod 644 /usr/share/heartbeat/heartbeat.yml && \
    chown heartbeat:heartbeat /usr/share/heartbeat/heartbeat.yml

USER heartbeat
