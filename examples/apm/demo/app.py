#!/usr/bin/env python3

import itertools
import json
import logging
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from opentelemetry import metrics, trace
from opentelemetry._logs import set_logger_provider
from opentelemetry.exporter.otlp.proto.grpc._log_exporter import OTLPLogExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.propagate import extract, inject
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.trace import SpanKind, Status, StatusCode


INVENTORY_URL = "http://demo-inventory:8080/stock"
CHECKOUT_URL = "http://demo-checkout:8080/checkout"
OTLP_ENDPOINT = "http://zabbix-proxy-sqlite3:4317"

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
)


def configure_telemetry(service_name):
    resource = Resource.create(
        {
            "service.name": service_name,
            "service.version": "1.0.0",
            "deployment.environment.name": "zabbix-apm-demo",
        }
    )

    tracer_provider = TracerProvider(resource=resource)
    tracer_provider.add_span_processor(
        BatchSpanProcessor(OTLPSpanExporter(endpoint=OTLP_ENDPOINT, insecure=True))
    )
    trace.set_tracer_provider(tracer_provider)
    tracer = trace.get_tracer("zabbix.apm.demo")

    metric_reader = PeriodicExportingMetricReader(
        OTLPMetricExporter(endpoint=OTLP_ENDPOINT, insecure=True),
        export_interval_millis=1000,
    )
    meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])
    metrics.set_meter_provider(meter_provider)
    meter = metrics.get_meter("zabbix.apm.demo")
    request_duration = meter.create_histogram(
        "demo.http.server.request.duration",
        unit="ms",
    )

    logger_provider = LoggerProvider(resource=resource)
    logger_provider.add_log_record_processor(
        BatchLogRecordProcessor(OTLPLogExporter(endpoint=OTLP_ENDPOINT, insecure=True))
    )
    set_logger_provider(logger_provider)
    logger = logging.getLogger("zabbix.apm.demo")
    logger.addHandler(LoggingHandler(level=logging.INFO, logger_provider=logger_provider))

    return tracer, request_duration, logger


def json_response(handler, status, payload):
    body = json.dumps(payload).encode()
    handler.send_response(status)
    handler.send_header("Content-Type", "application/json")
    handler.send_header("Content-Length", str(len(body)))
    handler.end_headers()
    handler.wfile.write(body)


def inventory_response(query):
    product_id = int(query["product"][0])
    time.sleep(0.02 + (product_id % 5) * 0.01)
    if product_id % 7 == 0:
        return 503, {"error": "inventory lookup failed", "product": product_id}

    return 200, {"product": product_id, "available": product_id % 3 != 0}


def checkout_response(query, tracer):
    order = int(query["order"][0])

    inventory_url = f"{INVENTORY_URL}?{urllib.parse.urlencode({'product': order})}"
    headers = {}
    with tracer.start_as_current_span(
        "GET inventory /stock",
        kind=SpanKind.CLIENT,
        attributes={
            "http.request.method": "GET",
            "server.address": "demo-inventory",
            "server.port": 8080,
            "url.full": inventory_url,
        },
    ) as span:
        inject(headers)
        try:
            request = urllib.request.Request(inventory_url, headers=headers)
            with urllib.request.urlopen(request, timeout=3) as response:
                status = response.status
                inventory = json.load(response)
        except urllib.error.HTTPError as error:
            status = error.code
            inventory = json.loads(error.read().decode())
        except OSError as error:
            status = 503
            inventory = {"error": str(error)}

        span.set_attribute("http.response.status_code", status)
        if status >= 500:
            span.set_status(Status(StatusCode.ERROR, inventory.get("error", "error")))

    return status, {"order": order, "inventory": inventory}


def serve():
    service_name = os.environ["SERVICE_NAME"]
    service_role = os.environ["SERVICE_ROLE"]
    service_route = {"inventory": "/stock", "checkout": "/checkout"}.get(service_role)
    tracer, request_duration, logger = configure_telemetry(service_name)

    class DemoHandler(BaseHTTPRequestHandler):
        def do_GET(self):
            parsed = urllib.parse.urlparse(self.path)
            route = parsed.path
            if route == "/health":
                json_response(self, 200, {"status": "ok"})
                return
            if route != service_route:
                json_response(self, 404, {"error": "not found"})
                return

            started = time.monotonic()
            parent_context = extract(self.headers)
            attributes = {
                "http.request.method": "GET",
                "http.route": route,
                "server.address": service_name,
                "server.port": 8080,
            }
            with tracer.start_as_current_span(
                f"GET {route}",
                context=parent_context,
                kind=SpanKind.SERVER,
                attributes=attributes,
            ) as span:
                query = urllib.parse.parse_qs(parsed.query)
                if route == "/stock":
                    status, payload = inventory_response(query)
                else:
                    status, payload = checkout_response(query, tracer)
                span.set_attribute("http.response.status_code", status)
                if status >= 500:
                    span.set_status(Status(StatusCode.ERROR, payload.get("error", "error")))

                metric_attributes = {
                    "http.request.method": "GET",
                    "http.route": route,
                    "http.response.status_code": status,
                }
                request_duration.record(
                    (time.monotonic() - started) * 1000,
                    metric_attributes,
                )
                log = logger.error if status >= 500 else logger.info
                log("GET %s completed with status %d", route, status)
                json_response(self, status, payload)

        def log_message(self, _format, *_args):
            return

    server = ThreadingHTTPServer(("0.0.0.0", 8080), DemoHandler)
    logger.info("%s listening on port 8080", service_name)
    server.serve_forever()


def load():
    for order in itertools.count(1):
        url = f"{CHECKOUT_URL}?{urllib.parse.urlencode({'order': order})}"
        try:
            with urllib.request.urlopen(url, timeout=5) as response:
                response.read()
        except OSError:
            pass
        time.sleep(1)


if __name__ == "__main__":
    {"serve": serve, "load": load}[sys.argv[1]]()
