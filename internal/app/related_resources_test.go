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

func TestMissingResourceList2(t *testing.T) {
	newClient(t).expect(http.MethodGet, "/employees/999/services", nil, http.StatusNotFound)
}

func TestMissingResourceList3(t *testing.T) {
	newClient(t).expect(http.MethodGet, "/services/999/employees", nil, http.StatusNotFound)
}
