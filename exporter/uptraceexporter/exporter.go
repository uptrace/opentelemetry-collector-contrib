// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uptraceexporter

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/uptrace/uptrace-go/spanexp"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/otel/label"
)

type traceExporter struct {
	cfg    *Config
	logger *zap.Logger
	upexp  *spanexp.Exporter
}

func newTraceExporter(cfg *Config, logger *zap.Logger) (*traceExporter, error) {
	if cfg.MaxBatchSize <= 0 {
		return nil, fmt.Errorf("uptrace: got batch_size=%d, wanted > 0", cfg.MaxBatchSize)
	}

	upexp, err := spanexp.NewExporter(&spanexp.Config{
		DSN: cfg.DSN,
	})
	if err != nil {
		return nil, err
	}

	exporter := &traceExporter{
		cfg:    cfg,
		logger: logger,
		upexp:  upexp,
	}

	return exporter, nil
}

// pushTraceData is the method called when trace data is available.
func (e *traceExporter) pushTraceData(ctx context.Context, td pdata.Traces) (int, error) {
	var outSpans []spanexp.Span

	rsSpans := td.ResourceSpans()
	for i := 0; i < rsSpans.Len(); i++ {
		rsSpan := rsSpans.At(i)
		resource := keyValueSlice(rsSpan.Resource().Attributes())

		ils := rsSpan.InstrumentationLibrarySpans()
		for j := 0; j < ils.Len(); j++ {
			ilsSpan := ils.At(j)
			lib := ilsSpan.InstrumentationLibrary()

			spans := ilsSpan.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				outSpans = append(outSpans, spanexp.Span{})
				out := &outSpans[len(outSpans)-1]

				out.ID = asUint64(span.SpanID().Bytes())
				out.ParentID = asUint64(span.ParentSpanID().Bytes())
				out.TraceID = span.TraceID().Bytes()

				out.Name = span.Name()
				out.Kind = spanKind(span.Kind())
				out.StartTime = int64(span.StartTime())
				out.EndTime = int64(span.EndTime())

				out.Resource = resource
				out.Attrs = keyValueSlice(span.Attributes())

				out.StatusCode = statusCode(span.Status())
				out.StatusMessage = span.Status().Message()

				out.Events = uptraceEvents(span.Events())
				out.Links = uptraceLinks(span.Links())

				out.Tracer.Name = lib.Name()
				out.Tracer.Version = lib.Version()

				if len(outSpans) >= e.cfg.MaxBatchSize {
					e.sendSpans(ctx, outSpans)
					outSpans = nil
				}
			}
		}
	}

	if len(outSpans) > 0 {
		e.sendSpans(ctx, outSpans)
	}

	return 0, nil
}

func (e *traceExporter) sendSpans(ctx context.Context, spans []spanexp.Span) {
	if err := e.upexp.SendSpans(ctx, spans); err != nil {
		e.logger.Warn(err.Error())
	}
}

func (e *traceExporter) Shutdown(ctx context.Context) error {
	return e.upexp.Shutdown(ctx)
}

func asUint64(b [8]byte) uint64 {
	return binary.LittleEndian.Uint64(b[:])
}

func spanKind(kind pdata.SpanKind) string {
	switch kind {
	case pdata.SpanKindCLIENT:
		return "client"
	case pdata.SpanKindSERVER:
		return "server"
	case pdata.SpanKindPRODUCER:
		return "producer"
	case pdata.SpanKindCONSUMER:
		return "consumer"
	case pdata.SpanKindINTERNAL:
		return "internal"
	case pdata.SpanKindUNSPECIFIED:
		fallthrough
	default:
		return "internal"
	}
}

func statusCode(status pdata.SpanStatus) string {
	switch status.Code() {
	case pdata.StatusCodeUnset:
		return "unset"
	case pdata.StatusCodeOk:
		return "ok"
	case pdata.StatusCodeError:
		return "error"
	}
	return "unset"
}

func keyValueSlice(attrs pdata.AttributeMap) spanexp.KeyValueSlice {
	out := make(spanexp.KeyValueSlice, 0, attrs.Len())

	attrs.ForEach(func(key string, value pdata.AttributeValue) {
		switch value.Type() {
		case pdata.AttributeValueSTRING:
			out = append(out, label.String(key, value.StringVal()))
		case pdata.AttributeValueBOOL:
			out = append(out, label.Bool(key, value.BoolVal()))
		case pdata.AttributeValueINT:
			out = append(out, label.Int64(key, value.IntVal()))
		case pdata.AttributeValueDOUBLE:
			out = append(out, label.Float64(key, value.DoubleVal()))
		case pdata.AttributeValueMAP:
			// TODO
		case pdata.AttributeValueARRAY:
			// TODO
		}
	})

	return out
}

func uptraceEvents(events pdata.SpanEventSlice) []spanexp.Event {
	if events.Len() == 0 {
		return nil
	}

	outEvents := make([]spanexp.Event, events.Len())
	for i := 0; i < events.Len(); i++ {
		event := events.At(i)

		out := &outEvents[i]
		out.Name = event.Name()
		out.Attrs = keyValueSlice(event.Attributes())
		out.Time = int64(event.Timestamp())
	}
	return outEvents
}

func uptraceLinks(links pdata.SpanLinkSlice) []spanexp.Link {
	if links.Len() == 0 {
		return nil
	}

	outLinks := make([]spanexp.Link, links.Len())
	for i := 0; i < links.Len(); i++ {
		link := links.At(i)

		out := &outLinks[i]
		out.TraceID = link.TraceID().Bytes()
		out.SpanID = asUint64(link.SpanID().Bytes())
		out.Attrs = keyValueSlice(link.Attributes())
	}
	return outLinks
}
