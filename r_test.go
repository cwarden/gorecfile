package recfile

import (
	"io"
	"strings"
	"testing"
)

func TestReaderSkipsDescriptors(t *testing.T) {
	input := `%rec: Task
%key: ID

ID: T1
Name: Write spec
EstimateLow: 4
EstimateHigh: 8

%rec: Task
ID: T2
Name: Implement feature
EstimateLow: 6
EstimateHigh: 10
DependsOn: T1

# A comment
ID: T3
Name: Testing
EstimateLow: 2
EstimateHigh: 4
DependsOn: T2

`

	r := NewReader(strings.NewReader(input))

	var records []map[string][]string
	for {
		m, err := r.NextMapWithSlice()
		if err != nil {
			break
		}
		records = append(records, m)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 task records, got %d", len(records))
	}

	if records[0]["ID"][0] != "T1" {
		t.Errorf("expected first ID to be T1, got %s", records[0]["ID"][0])
	}

	if records[1]["DependsOn"][0] != "T1" {
		t.Errorf("expected T2 to depend on T1")
	}

	if records[2]["EstimateHigh"][0] != "4" {
		t.Errorf("expected T3 EstimateHigh to be 4")
	}
}

func TestReaderWithRecordTypePersistence(t *testing.T) {
	const input = `
%rec: Package
%key: ID

ID: P1
Name: Core Features

ID: ASAP
Name: ASAP

%rec: Task
%key: ID

ID: T1
Name: Design
EstimateLow: 4
EstimateHigh: 8
Package: P1

ID: T2
Name: Implement
EstimateLow: 6
EstimateHigh: 10
DependsOn: T1
Package: P1
`
	r := NewReader(strings.NewReader(input))

	var packageCount, taskCount int
	var current string

	for {
		rec, err := r.NextMap()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("error reading record: %v", err)
		}

		recType := rec["%rec"]
		switch recType {
		case "Package":
			packageCount++
		case "Task":
			taskCount++
		default:
			t.Fatalf("unexpected record type: %q", recType)
		}

		if rec["ID"] == "" {
			t.Errorf("record is missing ID field: %+v", rec)
		}

		current = recType
	}

	if packageCount != 2 {
		t.Errorf("expected 2 package records, got %d", packageCount)
	}
	if taskCount != 2 {
		t.Errorf("expected 2 task records, got %d", taskCount)
	}
	if current != "Task" {
		t.Errorf("last record type should be Task, got %s", current)
	}
}
