package node

import (
	"path/filepath"
	"testing"

	"github.com/spacemeshos/go-spacemesh/log/logtest"
	"github.com/stretchr/testify/require"
)

func TestLoadEdSigners(t *testing.T) {
	tempdir := t.TempDir()
	app := New(WithLog(logtest.New(t)))
	app.Config.SMESHING.Opts.DataDir = filepath.Join(tempdir, "mesher1")
	signer1, err := app.LoadOrCreateEdSigner()
	app.Config.SMESHING.Opts.DataDir = filepath.Join(tempdir, "mesher2")
	signer2, err := app.LoadOrCreateEdSigner()
	app.Config.SMESHING.Opts.DataDir = filepath.Join(tempdir, "mesher3")
	signer3, err := app.LoadOrCreateEdSigner()

	app.Config.SMESHING.Opts.DataDir = tempdir
	_, err = app.LoadOrCreateEdSigner()

	signers, err := app.LoadEdSigners(tempdir)

	require.NoError(t, err)
	require.Equal(t, len(signers), 3)
	require.Equal(t, signers[signer1.NodeID()].edSgn, signer1)
	require.Equal(t, signers[signer2.NodeID()].edSgn, signer2)
	require.Equal(t, signers[signer3.NodeID()].edSgn, signer3)
	require.Equal(t, signers[signer1.NodeID()].path, filepath.Join(tempdir, "mesher1"))
	require.Equal(t, signers[signer2.NodeID()].path, filepath.Join(tempdir, "mesher2"))
	require.Equal(t, signers[signer3.NodeID()].path, filepath.Join(tempdir, "mesher3"))
}

func TestLoadEdSigner(t *testing.T) {
	tempdir := t.TempDir()
	app := New(WithLog(logtest.New(t)))
	app.Config.SMESHING.Opts.DataDir = tempdir
	signer, err := app.LoadOrCreateEdSigner()
	wrapsigner, err := app.LoadEdSigner(tempdir)
	require.NoError(t, err)
	require.Equal(t, signer, wrapsigner.edSgn)
	require.Equal(t, tempdir, wrapsigner.path)
}
