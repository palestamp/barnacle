package backends

import (
	"errors"
	"strings"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/xeipuuv/gojsonschema"
)

type SchemaValidationMiddleware struct {
	validator api.OptionsValidator
	backend   api.Backend
}

type JSONSchemaValidator struct {
	schema *gojsonschema.Schema
}

func (v *JSONSchemaValidator) Validate(ops api.QueueOptions) error {
	result, err := v.schema.Validate(gojsonschema.NewGoLoader(ops))
	if err != nil {
		return err
	}

	if !result.Valid() {
		return toError(result.Errors())
	}
	return nil
}

func toError(errs []gojsonschema.ResultError) error {
	var out []string
	for _, e := range errs {
		out = append(out, e.String())
	}

	return errors.New(strings.Join(out, "; "))
}

func NewJSONSchemaValidator(loader gojsonschema.JSONLoader) (*JSONSchemaValidator, error) {
	schema, err := gojsonschema.NewSchema(loader)
	return &JSONSchemaValidator{
		schema: schema,
	}, err
}

func NewValidationMiddleware(backend api.Backend, validator api.OptionsValidator) *SchemaValidationMiddleware {
	return &SchemaValidationMiddleware{
		validator: validator,
		backend:   backend,
	}
}

// Create forwards Create call to actual backend instance.
func (m *SchemaValidationMiddleware) Create(qm api.QueueMetadata) error {
	if err := qm.Options.Validate(m.validator); err != nil {
		return err
	}

	return m.backend.Create(qm)
}

// Connect forwards Connect call to actual backend instance.s
func (m *SchemaValidationMiddleware) Connect(qm api.QueueMetadata) (api.Queue, error) {
	if err := qm.Options.Validate(m.validator); err != nil {
		return nil, err
	}

	return m.backend.Connect(qm)
}
