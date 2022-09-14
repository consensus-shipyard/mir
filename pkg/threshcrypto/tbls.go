package threshcrypto

import (
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/drand/kyber"
	bls12381 "github.com/drand/kyber-bls12381"
	"github.com/drand/kyber/pairing"
	"github.com/drand/kyber/share"
	"github.com/drand/kyber/sign"
	"github.com/drand/kyber/sign/tbls"
	t "github.com/filecoin-project/mir/pkg/types"
)

type TBLSInst struct {
	t         int
	n         int
	scheme    sign.ThresholdScheme
	privShare *share.PriShare
	public    *share.PubPoly
}

func tbls12381Scheme() (pairing.Suite, sign.ThresholdScheme, kyber.Group, kyber.Group) {
	suite := bls12381.NewBLS12381Suite()
	scheme := tbls.NewThresholdSchemeOnG1(suite)
	sigGroup := suite.G1()
	keyGroup := suite.G2()

	return suite, scheme, sigGroup, keyGroup
}

func TBLS12381Keygen(t, n int, randSource cipher.Stream) ([]*TBLSInst, error) {
	if !(t >= (n+1)/2 && n > 0) {
		return nil, fmt.Errorf("TBLS requires t >= (n+1)/2 and n > 0")
	}

	pairing, scheme, sigGroup, keyGroup := tbls12381Scheme()

	if randSource == nil {
		randSource = pairing.RandomStream()
	}

	secret := sigGroup.Scalar().Pick(randSource)
	privFull := share.NewPriPoly(keyGroup, t, secret, randSource)
	public := privFull.Commit(keyGroup.Point().Base())

	privShares := privFull.Shares(n)
	instances := make([]*TBLSInst, n)
	for i := 0; i < n; i++ {
		instances[i] = &TBLSInst{
			scheme:    scheme,
			privShare: privShares[i],
			public:    public,
			t:         t,
			n:         n,
		}
	}

	return instances, nil
}

func (t *TBLSInst) MarshalTo(w io.Writer) (int, error) {
	written := 0

	marshalInt := func(v int) error {
		if err := binary.Write(w, binary.BigEndian, int64(v)); err != nil {
			return err
		}
		written += binary.Size(int64(v))
		return nil
	}

	marshalKyber := func(v kyber.Marshaling) error {
		n, err := v.MarshalTo(w)
		written += n
		return err
	}

	if err := marshalInt(t.t); err != nil {
		return written, err
	}
	if err := marshalInt(t.n); err != nil {
		return written, err
	}

	pubPoint, pubCommitments := t.public.Info()
	if err := marshalKyber(pubPoint); err != nil {
		return written, err
	}

	if err := marshalInt(len(pubCommitments)); err != nil {
		return written, err
	}

	for _, commitment := range pubCommitments {
		if err := marshalKyber(commitment); err != nil {
			return written, err
		}
	}

	if err := marshalInt(t.privShare.I); err != nil {
		return written, err
	}

	if err := marshalKyber(t.privShare.V); err != nil {
		return written, err
	}

	return written, nil
}

func (t *TBLSInst) UnmarshalFrom(r io.Reader) (int, error) {
	read := 0

	t.privShare = &share.PriShare{}
	t.public = &share.PubPoly{}

	_, scheme, _, keyGroup := tbls12381Scheme()
	t.scheme = scheme

	unmarshalInt := func(v *int) error {
		var vI64 int64
		if err := binary.Read(r, binary.BigEndian, &vI64); err != nil {
			return err
		}
		read += binary.Size(vI64)
		*v = int(vI64)

		if int64(*v) != vI64 {
			return fmt.Errorf("loss of int precision during decode")
		}

		return nil
	}

	unmarshalKyber := func(v kyber.Marshaling) error {
		n, err := v.UnmarshalFrom(r)
		read += n
		return err
	}

	if err := unmarshalInt(&t.t); err != nil {
		return read, err
	}
	if err := unmarshalInt(&t.n); err != nil {
		return read, err
	}

	pubPoint := keyGroup.Point()
	if err := unmarshalKyber(pubPoint); err != nil {
		return read, err
	}

	var pubCommitmentsLen int
	if err := unmarshalInt(&pubCommitmentsLen); err != nil {
		return read, err
	}

	pubCommitments := make([]kyber.Point, pubCommitmentsLen)
	for i := 0; i < pubCommitmentsLen; i++ {
		pubCommitments[i] = keyGroup.Point()
		if err := unmarshalKyber(pubCommitments[i]); err != nil {
			return read, err
		}
	}

	t.public = share.NewPubPoly(keyGroup, pubPoint, pubCommitments)

	if err := unmarshalInt(&t.privShare.I); err != nil {
		return read, err
	}

	t.privShare.V = keyGroup.Scalar()
	if err := unmarshalKyber(t.privShare.V); err != nil {
		return read, err
	}

	return read, nil
}

func (t *TBLSInst) SignShare(msg [][]byte) ([]byte, error) {
	return t.scheme.Sign(t.privShare, digest(msg))
}

func (t *TBLSInst) VerifyShare(msg [][]byte, sigShare []byte, nodeID t.NodeID) error {
	return t.scheme.VerifyPartial(t.public, digest(msg), sigShare)
}

func (t *TBLSInst) VerifyFull(msg [][]byte, sigFull []byte) error {
	return t.scheme.VerifyRecovered(t.public.Commit(), digest(msg), sigFull)
}

func (t *TBLSInst) Recover(msg [][]byte, sigShares [][]byte) ([]byte, error) {
	return t.scheme.Recover(t.public, digest(msg), sigShares, t.t, t.n)
}

func digest(data [][]byte) []byte {
	h := sha256.New()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}
