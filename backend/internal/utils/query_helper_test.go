package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func queryTestContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}

func TestParseQueryParamsMapsAllowedFields(t *testing.T) {
	policy := QueryPolicy{
		SortableFields:   map[string]string{"created": "created_at"},
		FilterableFields: map[string]string{"name": "name"},
		AllowedPreloads:  map[string]string{"items": "Items"},
	}
	params, err := ParseQueryParams(queryTestContext("/?orderBy=created&name=test&preloads=items"), policy)
	if err != nil {
		t.Fatal(err)
	}
	if params.OrderBy != "created_at" || params.Filters["name"] != "test" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if len(params.Preloads) != 1 || params.Preloads[0] != "Items" {
		t.Fatalf("unexpected preloads: %#v", params.Preloads)
	}
}

func TestParseQueryParamsRejectsUnknownSortField(t *testing.T) {
	_, err := ParseQueryParams(queryTestContext("/?orderBy=id%3BSELECT+1"), QueryPolicy{})
	if err == nil {
		t.Fatal("expected unsupported sort field error")
	}
}

func TestParseQueryParamsRejectsUnknownFilterField(t *testing.T) {
	_, err := ParseQueryParams(queryTestContext("/?id%3BSELECT+1=value"), QueryPolicy{})
	if err == nil {
		t.Fatal("expected unsupported filter field error")
	}
}

func TestParseQueryParamsRejectsUnknownPreload(t *testing.T) {
	_, err := ParseQueryParams(queryTestContext("/?preloads=Secrets"), QueryPolicy{})
	if err == nil {
		t.Fatal("expected unsupported preload error")
	}
}
