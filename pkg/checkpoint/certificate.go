package checkpoint

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// Certificate represents a certificate of validity of a checkpoint.
// It is included in a stable checkpoint itself.
type Certificate map[t.NodeID][]byte

func (cert *Certificate) Pb() map[string][]byte {
	return maputil.Transform(*cert,
		func(k t.NodeID, v []byte) (string, []byte) {
			return k.Pb(), v
		},
	)
}

func (cert *Certificate) Serialize() ([]byte, error) {
	em, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		return nil, err
	}

	return em.Marshal(cert)
}

func (cert *Certificate) Deserialize(data []byte) error {
	if err := cbor.Unmarshal(data, cert); err != nil {
		return fmt.Errorf("failed to CBOR unmarshal checkpoint certificate: %w", err)
	}
	return nil
}
