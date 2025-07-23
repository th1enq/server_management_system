FROM alpine

RUN apk add --no-cache bash curl jq

WORKDIR /app

COPY register.sh /app/register.sh
COPY send_metrics.sh /app/send_metrics.sh

RUN chmod +x /app/*.sh

CMD ["tail", "-f", "/dev/null"]
