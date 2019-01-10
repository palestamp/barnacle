package decoder

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func Decode(from interface{}, into interface{}) error {
	var m mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:   into,
		Metadata: &m,
	})
	if err != nil {
		return errors.WithMessage(err, "resource decoding failed")
	}

	if err := decoder.Decode(from); err != nil {
		return errors.WithMessage(err, "resource decoding failed")
	}

	if len(m.Unused) != 0 {
		err = fmt.Errorf("unknown fields: %s", strings.Join(m.Unused, ", "))
		return errors.WithMessage(err, "resource decoding failed")
	}

	return nil
}
