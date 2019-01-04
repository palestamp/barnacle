package backends_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/backends"
	"github.com/palestamp/barnacle/pkg/backends/postgres"
)

func TestConnector(t *testing.T) {
	registry := backends.NewRegistry()
	registry.RegisterConnector(api.BackendType("postgres"), &postgres.Connector{})

	connector, err := registry.Connector(api.BackendType("postgres"))
	assert.NoError(t, err)

	_, err = connector.Connect(api.ResourceConnOptions{
		"uri": "postgresql://postgres@localhost:5434/barnacle",
	})
	assert.NoError(t, err)

}
