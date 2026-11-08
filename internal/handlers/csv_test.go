package handlers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"reflect"
	"testing"
)

func TestCSVQuotingAndFormulaProtection(t *testing.T) {
	rows := [][]string{{"Анна, \"мастер\"", "line1\nline2", "=SUM(A1:A2)", "  @formula", "+7999", "-10", "\t=bad"}}
	var buf bytes.Buffer
	if err := encodeCSV(&buf, []string{"name", "notes", "formula", "spaces", "phone", "minus", "tab"}, rows); err != nil {
		t.Fatal(err)
	}
	got, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{rows[0][0], rows[0][1], "'=SUM(A1:A2)", "'  @formula", "'+7999", "'-10", "'\t=bad"}
	if !reflect.DeepEqual(got[1], want) {
		t.Fatalf("row: %#v", got[1])
	}
	if rows[0][2] != "=SUM(A1:A2)" {
		t.Fatal("input mutated")
	}
}

type failingCSVWriter struct{}

func (failingCSVWriter) Write([]byte) (int, error) { return 0, errors.New("disconnected") }
func TestCSVPropagatesWriteErrors(t *testing.T) {
	if err := encodeCSV(failingCSVWriter{}, []string{"id"}, nil); err == nil {
		t.Fatal("expected write failure")
	}
}
