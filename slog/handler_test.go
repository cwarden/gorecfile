// recfile -- GNU recutils'es recfiles parser on pure Go
// Copyright (C) 2020-2024 Sergey Matveev <stargrave@stargrave.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package slog

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"go.cypherpunks.su/recfile/v2"
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
	if !logger.Enabled(context.TODO(), slog.LevelWarn) {
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
