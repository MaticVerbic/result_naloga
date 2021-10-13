package docs

import (
	"github.com/pkg/errors"
	"github.com/swaggo/swag"
)

// InitSwag causes init() method to run and parses the doc file.
func InitSwag() ([]byte, error) {
	// required to force docs.init() method.
	SwaggerInfo.BasePath = "/"

	out, err := swag.ReadDoc()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read docs")
	}

	return []byte(out), nil
}
