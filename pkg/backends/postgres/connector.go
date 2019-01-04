package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/palestamp/barnacle/pkg/api"
)

type ResourceConnOptions struct {
	URI string `mapstructure:"uri"`
}

func OptionsFromResource(ops api.ResourceConnOptions) (*ResourceConnOptions, error) {
	var res ResourceConnOptions
	var m mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:   &res,
		Metadata: &m,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "resource decoding failed")
	}

	if err := decoder.Decode(ops); err != nil {
		return nil, errors.WithMessage(err, "resource decoding failed")
	}

	if len(m.Unused) != 0 {
		err = fmt.Errorf("unknown fields: %s", strings.Join(m.Unused, ", "))
		return nil, errors.WithMessage(err, "resource decoding failed")
	}

	return &res, nil
}

type Connector struct{}

func (c *Connector) Connect(ops api.ResourceConnOptions) (api.Backend, error) {
	op, err := OptionsFromResource(ops)
	if err != nil {
		return nil, err
	}

	connConfig, err := pgx.ParseURI(op.URI)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connConfig,
	})

	if err != nil {
		return nil, err
	}

	return NewBackendFromPool(pool), nil
}
