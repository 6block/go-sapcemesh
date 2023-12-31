// Code generated by github.com/spacemeshos/go-scale/scalegen. DO NOT EDIT.

// nolint
package multisig

import (
	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/common/types"
)

func (t *MultiSig) EncodeScale(enc *scale.Encoder) (total int, err error) {
	{
		n, err := scale.EncodeCompact8(enc, uint8(t.Required))
		if err != nil {
			return total, err
		}
		total += n
	}
	{
		n, err := scale.EncodeStructSliceWithLimit(enc, t.PublicKeys, 10)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (t *MultiSig) DecodeScale(dec *scale.Decoder) (total int, err error) {
	{
		field, n, err := scale.DecodeCompact8(dec)
		if err != nil {
			return total, err
		}
		total += n
		t.Required = uint8(field)
	}
	{
		field, n, err := scale.DecodeStructSliceWithLimit[types.Hash32](dec, 10)
		if err != nil {
			return total, err
		}
		total += n
		t.PublicKeys = field
	}
	return total, nil
}
