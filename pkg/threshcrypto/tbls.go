package threshcrypto

import (
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/drand/kyber"
	bls12381 "github.com/drand/kyber-bls12381"
	"github.com/drand/kyber/pairing"
	"github.com/drand/kyber/share"
	"github.com/drand/kyber/sign"
	"github.com/drand/kyber/sign/tbls"
	"golang.org/x/exp/slices"

	t "github.com/filecoin-project/mir/pkg/types"
)

type TBLSInst struct {
	t         int
	members   []t.NodeID
	scheme    sign.ThresholdScheme
	sigGroup  kyber.Group
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

func TBLS12381Keygen(T int, members []t.NodeID, randSource cipher.Stream) ([]*TBLSInst, error) {
	N := len(members)
	if !(T >= (N+1)/2 && N > 0) {
		return nil, fmt.Errorf("TBLS requires t >= (n+1)/2 and n > 0")
	}

	pairing, scheme, sigGroup, keyGroup := tbls12381Scheme()

	if randSource == nil {
		randSource = pairing.RandomStream()
	}

	secret := sigGroup.Scalar().Pick(randSource)
	privFull := share.NewPriPoly(keyGroup, T, secret, randSource)
	public := privFull.Commit(keyGroup.Point().Base())

	privShares := privFull.Shares(N)
	instances := make([]*TBLSInst, N)
	for i := 0; i < N; i++ {
		instances[i] = &TBLSInst{
			sigGroup:  sigGroup,
			scheme:    scheme,
			privShare: privShares[i],
			public:    public,
			t:         T,
			members:   members,
		}
	}

	return instances, nil
}

func (inst *TBLSInst) MarshalTo(w io.Writer) (int, error) {
	written := 0

	marshalInt := func(v int) error {
		if err := binary.Write(w, binary.BigEndian, int64(v)); err != nil {
			return err
		}
		written += binary.Size(int64(v))
		return nil
	}

	marshalString := func(v string) error {
		vBytes := []byte(v)

		if err := marshalInt(len(vBytes)); err != nil {
			return err
		}

		for _, b := range vBytes {
			if err := binary.Write(w, binary.BigEndian, b); err != nil {
				return err
			}
			written += binary.Size(b)
		}

		return nil
	}

	marshalKyber := func(v kyber.Marshaling) error {
		n, err := v.MarshalTo(w)
		written += n
		return err
	}

	if err := marshalInt(inst.t); err != nil {
		return written, err
	}

	if err := marshalInt(len(inst.members)); err != nil {
		return written, err
	}
	for _, member := range inst.members {
		if err := marshalString(string(member)); err != nil {
			return written, err
		}
	}

	pubPoint, pubCommitments := inst.public.Info()
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

	if err := marshalInt(inst.privShare.I); err != nil {
		return written, err
	}

	if err := marshalKyber(inst.privShare.V); err != nil {
		return written, err
	}

	return written, nil
}

func (inst *TBLSInst) UnmarshalFrom(r io.Reader) (int, error) {
	read := 0

	inst.privShare = &share.PriShare{}
	inst.public = &share.PubPoly{}

	_, scheme, sigGroup, keyGroup := tbls12381Scheme()
	inst.scheme = scheme
	inst.sigGroup = sigGroup

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

	unmarshalString := func() (string, error) {
		var size int
		if err := unmarshalInt(&size); err != nil {
			return "", err
		}

		strBytes := make([]byte, size)

		for i := range strBytes {
			if err := binary.Read(r, binary.BigEndian, &strBytes[i]); err != nil {
				return "", err
			}
			read += binary.Size(strBytes[i])
		}

		return string(strBytes), nil
	}

	unmarshalKyber := func(v kyber.Marshaling) error {
		n, err := v.UnmarshalFrom(r)
		read += n
		return err
	}

	if err := unmarshalInt(&inst.t); err != nil {
		return read, err
	}

	var nMembers int
	if err := unmarshalInt(&nMembers); err != nil {
		return read, err
	}

	members := make([]t.NodeID, nMembers)
	for i := range members {
		s, err := unmarshalString()
		if err != nil {
			return read, err
		}
		members[i] = t.NodeID(s)
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

	inst.public = share.NewPubPoly(keyGroup, pubPoint, pubCommitments)

	if err := unmarshalInt(&inst.privShare.I); err != nil {
		return read, err
	}

	inst.privShare.V = keyGroup.Scalar()
	if err := unmarshalKyber(inst.privShare.V); err != nil {
		return read, err
	}

	return read, nil
}

func (inst *TBLSInst) SignShare(msg [][]byte) ([]byte, error) {
	return inst.scheme.Sign(inst.privShare, digest(msg))
}

func (inst *TBLSInst) VerifyShare(msg [][]byte, sigShare []byte, nodeID t.NodeID) error {
	idx, err := tbls.SigShare(sigShare).Index()
	if err != nil {
		return err
	}

	if idx != slices.Index(inst.members, nodeID) {
		return fmt.Errorf("signature share belongs to another node")
	}

	return inst.scheme.VerifyPartial(inst.public, digest(msg), sigShare)
}

func (inst *TBLSInst) VerifyFull(msg [][]byte, sigFull []byte) error {
	return inst.scheme.VerifyRecovered(inst.public.Commit(), digest(msg), sigFull)
}

func (inst *TBLSInst) Recover(msg [][]byte, sigShares [][]byte) ([]byte, error) {
	// We don't use inst.scheme.Recover to avoid validating sigShares twice

	// This function is a modified version of the original implementation of inst.scheme.Recover
	// The original can be found at: https://github.com/drand/kyber/blob/9b6e107d216803c85237cd7c45196e5c545e447b/sign/tbls/tbls.go#L118

	var pubShares []*share.PubShare
	for _, sig := range sigShares {
		sh := tbls.SigShare(sig)
		i, err := sh.Index()
		if err != nil {
			continue
		}
		point := inst.sigGroup.Point()
		if err := point.UnmarshalBinary(sh.Value()); err != nil {
			continue
		}
		pubShares = append(pubShares, &share.PubShare{I: i, V: point})
		if len(pubShares) >= inst.t {
			break
		}
	}

	if len(pubShares) < inst.t {
		return nil, errors.New("not enough valid partial signatures")
	}

	commit, err := share.RecoverCommit(inst.sigGroup, pubShares, inst.t, len(inst.members))
	if err != nil {
		return nil, err
	}

	sig, err := commit.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func digest(data [][]byte) []byte {
	h := sha256.New()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}
