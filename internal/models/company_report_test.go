package models

import (
	"testing"
	"time"
)

func TestReportPeriodValidation(t *testing.T) {
	at := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	for _, q := range []ReportQuery{{}, {CompanyID: 1, From: at}, {CompanyID: 1, From: at, To: at}, {CompanyID: 1, From: at, To: at.Add(-time.Hour)}, {CompanyID: 1, From: at, To: at.Add(367 * 24 * time.Hour)}} {
		if err := q.Validate(); err == nil {
			t.Fatalf("accepted %+v", q)
		}
	}
	if err := (ReportQuery{CompanyID: 1, From: at, To: at.Add(366 * 24 * time.Hour)}).Validate(); err != nil {
		t.Fatal(err)
	}
}
