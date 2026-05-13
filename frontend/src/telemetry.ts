import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { FetchInstrumentation } from "@opentelemetry/instrumentation-fetch";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { BatchSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import {
  ATTR_SERVICE_NAME,
  ATTR_SERVICE_VERSION,
} from "@opentelemetry/semantic-conventions";

const DEFAULT_SERVICE_NAME = "pulse-check-frontend";

const collectorEndpoint =
  import.meta.env.VITE_OTEL_EXPORTER_OTLP_ENDPOINT?.trim();

if (collectorEndpoint) {
  const traceExporter = new OTLPTraceExporter({
    url: toTraceExportUrl(collectorEndpoint),
  });

  const tracerProvider = new WebTracerProvider({
    resource: resourceFromAttributes({
      [ATTR_SERVICE_NAME]:
        import.meta.env.VITE_OTEL_SERVICE_NAME ?? DEFAULT_SERVICE_NAME,
      [ATTR_SERVICE_VERSION]: import.meta.env.VITE_APP_VERSION ?? "0.1.0",
    }),
    spanProcessors: [new BatchSpanProcessor(traceExporter)],
  });

  tracerProvider.register();

  registerInstrumentations({
    instrumentations: [
      new FetchInstrumentation({
        clearTimingResources: true,
        propagateTraceHeaderCorsUrls: [/.*/],
      }),
    ],
  });
}

function toTraceExportUrl(endpoint: string): string {
  const normalizedEndpoint = endpoint.replace(/\/+$/, "");

  if (normalizedEndpoint.endsWith("/v1/traces")) {
    return normalizedEndpoint;
  }

  return `${normalizedEndpoint}/v1/traces`;
}
