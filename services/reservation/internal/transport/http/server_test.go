package httptransport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lab2/platform/rest"
)

const (
	testVersion = "test"
	docsPath    = "/docs"
	schemaPath  = "/openapi"
	apiPrefix   = "/testing"
)

func TestNewDocs(t *testing.T) {
	t.Parallel()
	router, api := rest.New(rest.Config{RequestTimeout: time.Second, APIPrefix: apiPrefix}, rest.OpenAPIConfig{
		Title: "Reservation", Version: testVersion, DocsPath: docsPath, SchemaPath: schemaPath,
	})
	if api.OpenAPI().Info.Title != "Reservation" {
		t.Fatal("OpenAPI title was not configured")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, apiPrefix+docsPath, nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "swagger-ui") {
		t.Fatalf("Swagger docs: status=%d body=%s", response.Code, response.Body.String())
	}
}
