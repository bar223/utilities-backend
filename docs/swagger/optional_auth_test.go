package swagger

import (
	"encoding/json"
	"testing"

	"github.com/swaggo/swag"
)

func TestOptionalAuthSecurity(t *testing.T) {
	raw, err := swag.ReadDoc("utilities")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/utilities", "/utilities/feed", "/utilities/feed/{id}"} {
		security := document.Paths[path]["get"].Security
		if len(security) != 2 {
			t.Fatalf("%s: expected token and guest alternatives", path)
		}
		if _, ok := security[0]["BearerAuth"]; !ok {
			t.Fatalf("%s: missing BearerAuth", path)
		}
		if len(security[1]) != 0 {
			t.Fatalf("%s: missing anonymous alternative", path)
		}
	}
	if len(document.Paths["/utilities/{id}"]["put"].Security) != 1 {
		t.Fatal("publication must still require a token")
	}
	if len(document.Paths["/users/login"]["post"].Security) != 0 {
		t.Fatal("login must not require a token")
	}
}
