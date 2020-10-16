// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"time"

	"github.com/pkg/errors"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/log"
	"github.com/hyperledger-labs/perun-node/session"
)

type node struct {
	log.Logger
	cfg      perun.NodeConfig
	sessions map[string]perun.SessionAPI
	psync.Mutex
}

// New returns a perun NodeAPI instance initialized using the given config.
// This should be called only once, subsequent calls after the first non error
// response will return an error.
func New(cfg perun.NodeConfig) (*node, error) {
	// Currently, credentials are required for initializing a blockchain backend that
	// can validate the deployed contracts. Only a session has credentials and not a node.
	// So for now, check only if the addresses can be parsed.
	// TODO: (mano) Implement a read-only blockchain backend that can be initialized without
	// credentials and use it here.
	wb := ethereum.NewWalletBackend()
	_, err := wb.ParseAddr(cfg.Adjudicator)
	if err != nil {
		return nil, errors.WithMessage(err, "default adjudicator address")
	}
	_, err = wb.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, errors.WithMessage(err, "default adjudicator address")
	}

	err = log.InitLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing logger for node")
	}
	return &node{
		Logger:   log.NewLoggerWithField("node", 1), // ID of the node is always 1.
		cfg:      cfg,
		sessions: make(map[string]perun.SessionAPI),
	}, nil
}

// Time returns the current UTC time as per the node's system clock in unix format.
func (n *node) Time() int64 {
	n.Debug("Received request: node.Time")
	return time.Now().UTC().Unix()
}

// GetConfig returns the node level configuration parameters.
func (n *node) GetConfig() perun.NodeConfig {
	n.Debug("Received request: node.GetConfig")
	return n.cfg
}

// Help returns the set of APIs offered by the node.
func (n *node) Help() []string {
	n.Debug("Received request: node.Help")
	return []string{"payment"}
}

// OpenSession opens a session on this node using the given config file and returns the
// session id, which can be used to retrieve a SessionAPI instance from this node instance.
func (n *node) OpenSession(configFile string) (string, []perun.ChInfo, error) {
	n.Debug("Received request: node.OpenSession", configFile)
	n.Lock()
	defer n.Unlock()

	sessionConfig, err := session.ParseConfig(configFile)
	if err != nil {
		n.Error(err, "parsing session config")
		return "", nil, perun.ErrInvalidConfig
	}
	sess, err := session.New(sessionConfig)
	if err != nil {
		n.Error(err, "initializing session")
		return "", nil, perun.GetAPIError(err)
	}
	n.sessions[sess.ID()] = sess
	return sess.ID(), sess.GetChsInfo(), nil
}

// GetSession is a special call that should be used internally to retrieve a SessionAPI
// instance to access its methods. This should not be exposed to the user.
func (n *node) GetSession(sessionID string) (perun.SessionAPI, error) {
	n.Debug("Internal call: GetSession", sessionID)
	n.Lock()
	defer n.Unlock()

	sess, ok := n.sessions[sessionID]
	if !ok {
		return nil, perun.ErrUnknownSessionID
	}
	return sess, nil
}