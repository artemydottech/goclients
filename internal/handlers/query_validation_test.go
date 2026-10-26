package handlers

import (
	"net/http"
	"testing"
)

func TestServicesRejectNonpositiveCompanyFilter(t *testing.T) {
	for _, value := range []string{"0", "-1", "99999999999999999999999"} {
		rec := serve(t, "GET /services", NewServiceHandler(nil).GetAllServices, http.MethodGet, "/services?company_id="+value, "")
		assertStatus(t, rec, http.StatusBadRequest)
	}
}
