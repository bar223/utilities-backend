package swagger

import (
	"encoding/json"

	"github.com/swaggo/swag"
)

// optionalAuthSpec preserves the generated documentation and adds the
// anonymous alternative for methods marked @x-optionalAuth true.
// The BearerAuth alternative makes Swagger UI send the authorized token;
// the empty alternative documents that guests can also call the method.
type optionalAuthSpec struct{}

func (optionalAuthSpec) ReadDoc() string {
	original := SwaggerInfo.ReadDoc()
	var document map[string]interface{}
	if err := json.Unmarshal([]byte(original), &document); err != nil {
		return original
	}
	paths, _ := document["paths"].(map[string]interface{})
	for _, path := range paths {
		operations, _ := path.(map[string]interface{})
		for _, value := range operations {
			operation, ok := value.(map[string]interface{})
			if ok && operation["x-optionalAuth"] == true {
				operation["security"] = []map[string][]string{{"BearerAuth": {}}, {}}
			}
		}
	}
	result, err := json.Marshal(document)
	if err != nil {
		return original
	}
	return string(result)
}

func init() {
	swag.Register("utilities", optionalAuthSpec{})
}
