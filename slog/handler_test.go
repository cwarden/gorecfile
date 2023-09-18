package slog

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"go.cypherpunks.ru/recfile"
)

func TestBasic(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewRecfileHandler(
		&buf,
		slog.LevelWarn,
		"Urgency",
		"Message",
		"Time",
	))
	if !logger.Enabled(nil, slog.LevelWarn) {
		t.FailNow()
	}
	logger.Info("won't catch me")
	logger.Warn("catch me")

	r := recfile.NewReader(&buf)
	m, err := r.NextMap()
	if err != nil {
		t.Fatal(err)
	}
	if m["Message"] != "catch me" {
		t.FailNow()
	}
	if m["Urgency"] != "WARN" {
		t.FailNow()
	}
	if _, err = time.Parse(time.RFC3339Nano, m["Time"]); err != nil {
		t.FailNow()
	}
}

func TestTrimmed(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewRecfileHandler(&buf, slog.LevelWarn, "", "Message", ""))
	logger.Warn("catch me")
	r := recfile.NewReader(&buf)
	m, err := r.NextMap()
	if err != nil {
		t.Fatal(err)
	}
	if m["Message"] != "catch me" {
		t.FailNow()
	}
	if m["Urgency"] != "" {
		t.FailNow()
	}
	if m["Time"] != "" {
		t.FailNow()
	}
}

func TestFeatured(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewRecfileHandler(&buf, slog.LevelInfo, "L", "M", "T"))
	logger.WithGroup("grou").WithGroup("py").With("foo", "bar").With("bar", "baz").Info(
		"catch me", "baz", []string{"multi", "line"},
	)
	r := recfile.NewReader(&buf)
	m, err := r.NextMap()
	if err != nil {
		t.Fatal(err)
	}
	if m["M"] != "catch me" {
		t.Fatal("M")
	}
	if m["L"] != "INFO" {
		t.Fatal("L")
	}
	if _, err = time.Parse(time.RFC3339Nano, m["T"]); err != nil {
		t.Fatal("T")
	}
	if m["grou_py_foo"] != "bar" {
		t.Fatal("foo")
	}
	if m["grou_py_bar"] != "baz" {
		t.Fatal("bar")
	}
	if m["grou_py_baz"] != "multi\nline" {
		t.Fatal("baz")
	}
}
