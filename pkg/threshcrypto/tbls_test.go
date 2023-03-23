package threshcrypto

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/types"
)

func TestTBLSHappySmoke(t *testing.T) {
	// confirm basic functionality is there (happy path)
	N := 5
	T := 3

	keys := keygen(T-1, T, N)

	data := [][]byte{{1, 2, 3, 4, 5}, {4, 2}}

	shares := make([][]byte, 3)
	for i := range shares {
		sh, err := keys[i].SignShare(data)
		assert.NoError(t, err)

		shares[i] = sh
	}

	// everyone can verify everyone's share
	for _, k := range keys {
		for i, sh := range shares {
			require.NoError(t, k.VerifyShare(data, sh, types.NewNodeIDFromInt(i)))
		}
	}

	for _, kR := range keys {
		// everyone can recover the full signature
		full, err := kR.Recover(data, shares)
		assert.NoError(t, err)

		// everyone can verify the recovered signature
		for _, kV := range keys {
			assert.NoError(t, kV.VerifyFull(data, full))
		}
	}
}

func TestTBLSSadSmoke(t *testing.T) {
	// confirm some basic problems are detected correctly
	N := 5
	T := 3

	keys := keygen(T-1, T, N)

	data := [][]byte{{1, 2, 3, 4, 5}, {4, 2}}

	shares := make([][]byte, 4)
	for i := range shares {
		sh, err := keys[i].SignShare(data)
		require.NoError(t, err)

		shares[i] = sh
	}

	// all of the same share is no good
	_, err := keys[0].Recover(data, [][]byte{shares[0], shares[0], shares[0]})
	assert.Error(t, err)

	// too little shares is no good
	_, err = keys[0].Recover(data, shares[:T-1])
	assert.Error(t, err)

	// mangle one of the shares
	shares[1][0] ^= 1

	// mangled share fails verification
	for _, k := range keys {
		assert.Error(t, k.VerifyShare(data, shares[1], "1"))
	}
}

func TestTBLSMarshalling(t *testing.T) {
	N := 3
	T := N
	nByz := (N+1)/2 - 1

	keys := keygen(nByz, T, N)

	keys2 := marshalUnmarshalKeys(t, keys)

	data := [][]byte{{1, 2, 3, 4, 5}, {4, 2}}

	// produce all required shares using both sets of keys
	sigShares := make([][]byte, N)
	sigShares2 := make([][]byte, N)
	for i := range sigShares {
		var err error
		sigShares[i], err = keys[i].SignShare(data)
		assert.NoError(t, err)

		sigShares2[i], err = keys2[i].SignShare(data)
		assert.NoError(t, err)
	}

	// check that the both sets of keys can recover the other's signature
	for i := range sigShares {
		var err error

		_, err = keys[i].Recover(data, sigShares2)
		assert.NoError(t, err)

		_, err = keys2[i].Recover(data, sigShares)
		assert.NoError(t, err)
	}
}

func keygen(_, T, N int) []*TBLSInst {
	members := make([]types.NodeID, N)
	for i := range members {
		members[i] = types.NewNodeIDFromInt(i)
	}

	rand := pseudorandomStream(DefaultPseudoSeed)
	return TBLS12381Keygen(T, members, rand)
}

func marshalUnmarshalKeys(t *testing.T, src []*TBLSInst) []*TBLSInst {
	res := make([]*TBLSInst, len(src))

	pipeR, pipeW := io.Pipe()

	go func() {
		for i := range src {
			_, err := src[i].MarshalTo(pipeW)
			assert.NoError(t, err)
		}

		pipeW.Close()
	}()

	for i := range res {
		res[i] = &TBLSInst{}
		_, err := res[i].UnmarshalFrom(pipeR)
		assert.NoError(t, err)
	}

	data := make([]byte, 1)
	_, err := pipeR.Read(data)
	assert.ErrorIs(t, err, io.EOF)
	pipeR.Close()
	return res
}
