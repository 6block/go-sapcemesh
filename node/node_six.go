package node

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/signing"
)

type WrapEdSigner struct {
	edSgn *signing.EdSigner
	path  string
}

type WrapEdSignerSets map[types.NodeID]*WrapEdSigner

func (app *App) LoadEdSigners(dir string) (WrapEdSignerSets, error) {
	sets := make(map[types.NodeID]*WrapEdSigner)
	rootdir := dir
	filepath.WalkDir(rootdir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			// filter root dir
			if path == rootdir {
				return nil
			}
			filename := filepath.Join(path, edKeyFileName)
			app.log.Info("Looking for identity file at `%v`", filename)

			data, err := os.ReadFile(filename)
			if err != nil {
				app.log.Error("failed to read identity file: %w", err)
				return nil
			}
			dst := make([]byte, signing.PrivateKeySize)
			n, err := hex.Decode(dst, data)
			if err != nil {
				app.log.Error("decoding private key: %w", err)
				return nil
			}
			if n != signing.PrivateKeySize {
				app.log.Error("invalid key size %d/%d", n, signing.PrivateKeySize)
				return nil
			}
			edSgn, err := signing.NewEdSigner(
				signing.WithPrivateKey(dst),
				signing.WithPrefix(app.Config.Genesis.GenesisID().Bytes()),
			)
			if err != nil {
				app.log.Error("failed to construct identity from data file: %w", err)
				return nil
			}
			app.log.Info("Loaded existing identity; public key: %v at path %v", edSgn.PublicKey(), filename)
			sets[edSgn.NodeID()] = &WrapEdSigner{edSgn, path}
		}
		return nil
	})
	return sets, nil
}

func (app *App) LoadEdSigner(dir string) (*WrapEdSigner, error) {
	filename := filepath.Join(dir, edKeyFileName)
	app.log.Info("Looking for identity file at `%v`", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read iddentity file %w", err)
	}
	dst := make([]byte, signing.PrivateKeySize)
	n, err := hex.Decode(dst, data)
	if err != nil {
		return nil, fmt.Errorf("decoding private key: %w", err)
	}
	if n != signing.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size %d/%d", n, signing.PrivateKeySize)
	}
	edSgn, err := signing.NewEdSigner(
		signing.WithPrivateKey(dst),
		signing.WithPrefix(app.Config.Genesis.GenesisID().Bytes()),
	)
	if err != nil {
		return nil, fmt.Errorf("falied to construct identity from data file: %w", err)
	}
	return &WrapEdSigner{edSgn: edSgn, path: dir}, nil
}

func (signers *WrapEdSignerSets) VRFSigners() map[types.NodeID]*signing.VRFSigner {
	sets := make(map[types.NodeID]*signing.VRFSigner)
	for k, v := range *signers {
		sets[k] = v.edSgn.VRFSigner()
	}
	return sets
}
