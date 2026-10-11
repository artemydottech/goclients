package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestContextHandlerAddsRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil)))
	ctx := context.WithValue(context.Background(), requestIDKey{}, "req-42")

	logger.InfoContext(ctx, "something happened")

	if !strings.Contains(buf.String(), "request_id=req-42") {
		t.Errorf("log line %q has no request id", buf.String())
	}
}

func TestContextHandlerWithoutRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil)))

	logger.Info("background job")

	if strings.Contains(buf.String(), "request_id") {
		t.Errorf("log line %q must not contain a request id", buf.String())
	}
}

func TestContextHandlerSurvivesWith(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil))).With("component", "db")
	ctx := context.WithValue(context.Background(), requestIDKey{}, "req-7")

	logger.InfoContext(ctx, "query")

	line := buf.String()
	if !strings.Contains(line, "component=db") || !strings.Contains(line, "request_id=req-7") {
		t.Errorf("log line %q lost an attribute", line)
	}
}
