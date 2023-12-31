package activation

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/spacemeshos/poet/logging"
	"github.com/spacemeshos/poet/shared"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/signing"
)

func TestHTTPPoet(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	var eg errgroup.Group
	t.Cleanup(func() { r.NoError(eg.Wait()) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := NewHTTPPoetTestHarness(logging.NewContext(ctx, zaptest.NewLogger(t)), t.TempDir())
	r.NoError(err)
	r.NotNil(c)

	eg.Go(func() error {
		err := c.Service.Start(ctx)
		return errors.Join(err, c.Service.Close())
	})

	client, err := NewHTTPPoetClient(c.RestURL().String(), DefaultPoetConfig(), WithLogger(zaptest.NewLogger(t)))
	require.NoError(t, err)

	resp, err := client.PowParams(context.Background())
	r.NoError(err)

	signer, err := signing.NewEdSigner(signing.WithPrefix([]byte("prefix")))
	require.NoError(t, err)
	ch := types.RandomHash()

	nonce, err := shared.FindSubmitPowNonce(
		context.Background(),
		resp.Challenge,
		ch.Bytes(),
		signer.NodeID().Bytes(),
		uint(resp.Difficulty),
	)
	r.NoError(err)

	signature := signer.Sign(signing.POET, ch.Bytes())
	prefix := bytes.Join([][]byte{signer.Prefix(), {byte(signing.POET)}}, nil)

	poetRound, err := client.Submit(context.Background(), time.Time{}, prefix, ch.Bytes(), signature, signer.NodeID(), PoetPoW{
		Nonce:  nonce,
		Params: *resp,
	})
	r.NoError(err)
	r.NotNil(poetRound)
}

func TestSumbmitTooLate(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	var eg errgroup.Group
	t.Cleanup(func() { r.NoError(eg.Wait()) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := NewHTTPPoetTestHarness(logging.NewContext(ctx, zaptest.NewLogger(t)), t.TempDir())
	r.NoError(err)
	r.NotNil(c)

	eg.Go(func() error {
		err := c.Service.Start(ctx)
		return errors.Join(err, c.Service.Close())
	})

	client, err := NewHTTPPoetClient(c.RestURL().String(), DefaultPoetConfig(), WithLogger(zaptest.NewLogger(t)))
	require.NoError(t, err)

	resp, err := client.PowParams(context.Background())
	r.NoError(err)

	signer, err := signing.NewEdSigner(signing.WithPrefix([]byte("prefix")))
	require.NoError(t, err)
	ch := types.RandomHash()

	nonce, err := shared.FindSubmitPowNonce(
		context.Background(),
		resp.Challenge,
		ch.Bytes(),
		signer.NodeID().Bytes(),
		uint(resp.Difficulty),
	)
	r.NoError(err)

	signature := signer.Sign(signing.POET, ch.Bytes())
	prefix := bytes.Join([][]byte{signer.Prefix(), {byte(signing.POET)}}, nil)

	_, err = client.Submit(context.Background(), time.Now(), prefix, ch.Bytes(), signature, signer.NodeID(), PoetPoW{
		Nonce:  nonce,
		Params: *resp,
	})
	r.ErrorIs(err, ErrInvalidRequest)
}

func TestCheckRetry(t *testing.T) {
	t.Parallel()
	t.Run("doesn't retry on context cancellation.", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		retry, err := checkRetry(ctx, nil, nil)
		require.ErrorIs(t, err, context.Canceled)
		require.False(t, retry)
	})
	t.Run("doesn't retry on unrecoverable error.", func(t *testing.T) {
		t.Parallel()
		retry, err := checkRetry(context.Background(), nil, &url.Error{Err: errors.New("unsupported protocol scheme")})
		require.NoError(t, err)
		require.False(t, retry)
	})
	t.Run("retries on 404 (not found).", func(t *testing.T) {
		t.Parallel()
		retry, err := checkRetry(context.Background(), &http.Response{StatusCode: http.StatusNotFound}, nil)
		require.NoError(t, err)
		require.True(t, retry)
	})
}
