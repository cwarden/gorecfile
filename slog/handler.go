// recfile -- GNU recutils'es recfiles parser on pure Go
// Copyright (C) 2020-2025 Sergey Matveev <stargrave@stargrave.org>
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
	"io"
	"log/slog"
	"sync"
	"time"

	"go.cypherpunks.su/recfile/v2"
)

type RecfileHandler struct {
	W        io.Writer
	LevelKey string
	MsgKey   string
	TimeKey  string
	group    string
	m        *sync.Mutex
	attrs    []slog.Attr
	Level    slog.Level
}

func NewRecfileHandler(
	w io.Writer,
	level slog.Level,
	levelKey, msgKey, timeKey string,
) *RecfileHandler {
	return &RecfileHandler{
		W:        w,
		Level:    level,
		LevelKey: levelKey,
		MsgKey:   msgKey,
		TimeKey:  timeKey,
		m:        new(sync.Mutex),
	}
}

func (h *RecfileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.Level
}

func (h *RecfileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RecfileHandler{
		W:        h.W,
		Level:    h.Level,
		LevelKey: h.LevelKey,
		MsgKey:   h.MsgKey,
		TimeKey:  h.TimeKey,
		attrs:    append(h.attrs, attrs...),
		group:    h.group,
		m:        h.m,
	}
}

func (h *RecfileHandler) WithGroup(name string) slog.Handler {
	neu := RecfileHandler{
		W:        h.W,
		Level:    h.Level,
		LevelKey: h.LevelKey,
		MsgKey:   h.MsgKey,
		TimeKey:  h.TimeKey,
		attrs:    h.attrs,
		group:    h.group + name + "_",
		m:        h.m,
	}
	return &neu
}

func writeValue(w *recfile.Writer, group string, attr slog.Attr) (err error) {
	if attr.Value.Kind() == slog.KindAny {
		multiline, ok := attr.Value.Any().([]string)
		if ok {
			if len(multiline) > 0 {
				_, err = w.WriteFieldMultiline(group+attr.Key, multiline)
				return
			}
			return
		}
	}
	_, err = w.WriteFields(recfile.Field{
		Name:  group + attr.Key,
		Value: attr.Value.String(),
	})
	return
}

func (h *RecfileHandler) Handle(ctx context.Context, rec slog.Record) (err error) {
	var b bytes.Buffer
	w := recfile.NewWriter(&b)
	_, err = w.RecordStart()
	if err != nil {
		panic(err)
	}
	var fields []recfile.Field
	if h.LevelKey != "" {
		fields = append(fields, recfile.Field{
			Name:  h.LevelKey,
			Value: rec.Level.String(),
		})
	}
	if h.TimeKey != "" {
		fields = append(fields, recfile.Field{
			Name:  h.TimeKey,
			Value: rec.Time.UTC().Format(time.RFC3339Nano),
		})
	}
	fields = append(fields, recfile.Field{Name: h.MsgKey, Value: rec.Message})
	_, err = w.WriteFields(fields...)
	if err != nil {
		return
	}
	for _, attr := range h.attrs {
		writeValue(w, h.group, attr)
	}
	rec.Attrs(func(attr slog.Attr) bool { return writeValue(w, h.group, attr) == nil })
	h.m.Lock()
	n, err := h.W.Write(b.Bytes())
	h.m.Unlock()
	if err != nil {
		return
	}
	if n != b.Len() {
		return io.EOF
	}
	return
}
