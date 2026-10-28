package app

import (
	"net/http"
	"testing"
)

func TestMissingResourceList0(t *testing.T) {
	newClient(t).expect(http.MethodGet, "/employees/999/time-off", nil, http.StatusNotFound)
}

func TestMissingResourceList1(t *testing.T) {
	newClient(t).expect(http.MethodGet, "/employees/999/schedule", nil, http.StatusNotFound)
}
