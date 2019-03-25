package modules

import "io"

const (
	// BlockCreatorDir is the name of the directory that is used to store the BlockCreator's
	// persistent data.
	BlockCreatorDir = "blockcreator"
)

// The BlockCreator interface provides access to BlockCreator features.
type BlockCreator interface {
	io.Closer
}
