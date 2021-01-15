# OpenTelemetry Collector configured with Uptrace exporter

This example demonstrates how to use OpenTelemetry Collector with Uptrace. Before running it you need to specify Uptrace DSN in `otel-collector-config.yml`:

```yaml
exporters:
  uptrace:
    dsn: "https://<token>@api.uptrace.dev/<project_id>"
```

Then build and start Otel Collector:

```shell
docker-compose up
```
