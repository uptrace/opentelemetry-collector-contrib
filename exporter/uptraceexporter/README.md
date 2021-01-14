# Uptrace Exporter

This exporter supports sending trace data to [Uptrace.dev](https://uptrace.dev).

## Configuration

| Configuration option |          | Note                                               |
| -------------------- | -------- | -------------------------------------------------- |
| `dsn`                | required | Data source name for your Uptrace project.         |
| `max_batch_size`     | optional | Maximum number of spans to send in a single batch. |

Example:

```yaml
exporters:
  uptrace:
    dsn: "https://<key>@api.uptrace.dev/<project_id>"
```
