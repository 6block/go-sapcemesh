package hare

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/go-spacemesh/codec"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log/logtest"
	"github.com/spacemeshos/go-spacemesh/signing"
)

func marshallUnmarshall(t *testing.T, msg *Message) *Message {
	buf, err := codec.Encode(msg)
	require.NoError(t, err)

	m := &Message{}
	require.NoError(t, codec.Decode(buf, m))
	return m
}

func TestBuilder_TestBuild(t *testing.T) {
	b := newMessageBuilder()
	signer, err := signing.NewEdSigner()
	require.NoError(t, err)
	msg := b.SetLayer(instanceID1).Sign(signer).Build()

	m := marshallUnmarshall(t, msg)
	assert.Equal(t, m, msg)
}

func TestMessageBuilder_SetValues(t *testing.T) {
	s := NewSetFromValues(types.ProposalID{5})
	msg := newMessageBuilder().SetValues(s).Build()

	m := marshallUnmarshall(t, msg)
	s1 := NewSet(m.Values)
	s2 := NewSet(msg.Values)
	assert.True(t, s1.Equals(s2))
}

func TestMessageBuilder_SetCertificate(t *testing.T) {
	signer, err := signing.NewEdSigner()
	require.NoError(t, err)

	s := NewSetFromValues(types.ProposalID{5})
	et := NewEligibilityTracker(1)
	tr := newCommitTracker(logtest.New(t), commitRound, make(chan *types.MalfeasanceGossip), et, 1, 1, s)
	m := BuildCommitMsg(signer, s)
	et.Track(m.SmesherID, m.Round, m.Eligibility.Count, true)
	tr.OnCommit(context.Background(), m)
	cert := tr.BuildCertificate()
	assert.NotNil(t, cert)
	c := newMessageBuilder().SetCertificate(cert).Build()
	cert2 := marshallUnmarshall(t, c).Cert
	assert.Equal(t, cert.Values, cert2.Values)
}

func TestMessageFromBuffer(t *testing.T) {
	b := newMessageBuilder()
	signer, err := signing.NewEdSigner()
	require.NoError(t, err)
	msg := b.SetLayer(instanceID1).Sign(signer).Build()

	buf, err := codec.Encode(msg)
	require.NoError(t, err)

	got, err := MessageFromBuffer(buf)
	require.NoError(t, err)
	require.Equal(t, msg, got)
}
