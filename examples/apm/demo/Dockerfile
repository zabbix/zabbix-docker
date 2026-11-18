FROM python:3.14-alpine

ARG OPENTELEMETRY_VERSION=1.45.1

RUN pip install --no-cache-dir \
        "opentelemetry-api==${OPENTELEMETRY_VERSION}" \
        "opentelemetry-exporter-otlp-proto-grpc==${OPENTELEMETRY_VERSION}" \
        "opentelemetry-sdk==${OPENTELEMETRY_VERSION}"

WORKDIR /app

COPY app.py .

USER 65534

EXPOSE 8080

ENTRYPOINT ["python", "/app/app.py"]
CMD ["serve"]
