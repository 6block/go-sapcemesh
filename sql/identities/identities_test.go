package identities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/go-spacemesh/codec"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
)

func TestMalicious(t *testing.T) {
	db := sql.InMemory()

	nodeID := types.NodeID{1, 1, 1, 1}
	mal, err := IsMalicious(db, nodeID)
	require.NoError(t, err)
	require.False(t, mal)

	var ballotProof types.BallotProof
	for i := 0; i < 2; i++ {
		ballotProof.Messages[i] = types.BallotProofMsg{
			InnerMsg: types.BallotMetadata{
				Layer:   types.LayerID(9),
				MsgHash: types.RandomHash(),
			},
			Signature: types.RandomEdSignature(),
			SmesherID: nodeID,
		}
	}
	proof := &types.MalfeasanceProof{
		Layer: types.LayerID(11),
		Proof: types.Proof{
			Type: types.MultipleBallots,
			Data: &ballotProof,
		},
	}
	now := time.Now()
	data, err := codec.Encode(proof)
	require.NoError(t, err)
	require.NoError(t, SetMalicious(db, nodeID, data, now))

	mal, err = IsMalicious(db, nodeID)
	require.NoError(t, err)
	require.True(t, mal)

	got, err := GetMalfeasanceProof(db, nodeID)
	require.NoError(t, err)
	require.Equal(t, now.UTC(), got.Received().UTC())
	got.SetReceived(time.Time{})
	require.EqualValues(t, proof, got)
}

func Test_GetMalicious(t *testing.T) {
	db := sql.InMemory()
	got, err := GetMalicious(db)
	require.NoError(t, err)
	require.Nil(t, got)

	const numBad = 11
	bad := make([]types.NodeID, 0, numBad)
	for i := 0; i < numBad; i++ {
		nid := types.NodeID{byte(i + 1)}
		bad = append(bad, nid)
		require.NoError(t, SetMalicious(db, nid, types.RandomBytes(11), time.Now().Local()))
	}
	got, err = GetMalicious(db)
	require.NoError(t, err)
	require.Equal(t, bad, got)
}
