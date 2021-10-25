// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package uploadselection

import (
	"storj.io/common/storj"
)

// Node defines necessary information for node-selection.
type Node struct {
	storj.NodeURL
	LastNet    string
	LastIPPort string
}

// Clone returns a deep clone of the selected node.
func (node *Node) Clone() *Node {
	return &Node{
		NodeURL:    node.NodeURL,
		LastNet:    node.LastNet,
		LastIPPort: node.LastIPPort,
	}
}
