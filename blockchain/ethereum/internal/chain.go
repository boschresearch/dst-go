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

package internal

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/pkg/errors"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node/blockchain"
)

// ChainBackend provides ethereum specific contract backend functionality.
type ChainBackend struct {
	// Cb is the instance of contract backend that will be used for all on-chain communications.
	Cb *pethchannel.ContractBackend
	// TxTimeout is the max time to wait for confirmation of transactions on blockchain.
	// If this expires, a transactions is considered failed.
	// Use sufficiently large values when connecting to mainnet.
	TxTimeout time.Duration
}

// NewFunder initializes and returns an instance of ethereum funder.
func (cb *ChainBackend) NewFunder(assetAddr pwallet.Address, onChainAddr pwallet.Address) pchannel.Funder {
	asset := pethwallet.AsWalletAddr(pethwallet.AsEthAddr(assetAddr))
	onChainAcc := accounts.Account{Address: pethwallet.AsEthAddr(onChainAddr)}
	onChainAccs := map[pethchannel.Asset]accounts.Account{*asset: onChainAcc}
	depositors := map[pethchannel.Asset]pethchannel.Depositor{*asset: new(pethchannel.ETHDepositor)}
	return pethchannel.NewFunder(*cb.Cb, onChainAccs, depositors)
}

// NewAdjudicator initializes and returns an instance of ethereum adjudicator.
func (cb *ChainBackend) NewAdjudicator(adjAddr, onChainAddr pwallet.Address) pchannel.Adjudicator {
	onChainAcc := accounts.Account{Address: pethwallet.AsEthAddr(onChainAddr)}
	return pethchannel.NewAdjudicator(*cb.Cb, pethwallet.AsEthAddr(adjAddr),
		pethwallet.AsEthAddr(onChainAddr), onChainAcc)
}

// ValidateAdjudicator validates the integrity of adjudicator contract at the
// given address.
func (cb *ChainBackend) ValidateAdjudicator(adjAddr pwallet.Address) error {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()
	err := pethchannel.ValidateAdjudicator(ctx, *cb.Cb, pethwallet.AsEthAddr(adjAddr))
	if pethchannel.IsErrInvalidContractCode(err) {
		return blockchain.NewInvalidContractError(blockchain.Adjudicator, adjAddr.String(), err)
	}
	return errors.Wrap(err, "validating adjudicator contract")
}

// ValidateAssetHolderETH validates the integrity of adjudicator and asset
// holder contracts at the given addresses.
//
// TODO: Submit a suggestion to go-perun to not validate the adjudicator contract in ValidateAssetHolder.
// If accepted, then update this function to
// validate only the asset holder contract.
func (cb *ChainBackend) ValidateAssetHolderETH(adjAddr, assetAddr pwallet.Address) error {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()
	// Though integrity of adjudicator is implicitly checked by ValidateAssetHolderETH,
	// we do it before that call to identify this type of error.
	err := pethchannel.ValidateAdjudicator(ctx, *cb.Cb, pethwallet.AsEthAddr(adjAddr))
	if pethchannel.IsErrInvalidContractCode(err) {
		return blockchain.NewInvalidContractError(blockchain.Adjudicator, adjAddr.String(), err)
	}

	err = pethchannel.ValidateAssetHolderETH(ctx, *cb.Cb, pethwallet.AsEthAddr(assetAddr), pethwallet.AsEthAddr(adjAddr))
	if pethchannel.IsErrInvalidContractCode(err) {
		return blockchain.NewInvalidContractError(blockchain.AssetHolderETH, assetAddr.String(), err)
	}
	return errors.Wrap(err, "validating asset holder ETH contract")
}

// DeployAdjudicator deploys the adjudicator contract.
func (cb *ChainBackend) DeployAdjudicator(onChainAddr pwallet.Address) (pwallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	onChainAcc := accounts.Account{Address: pethwallet.AsEthAddr(onChainAddr)}
	addr, err := pethchannel.DeployAdjudicator(ctx, *cb.Cb, onChainAcc)
	return pethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying adjudicator contract")
}

// DeployAsset deploys the asset holder contract, setting the adjudicator address to given value.
func (cb *ChainBackend) DeployAsset(adjAddr, onChainAddr pwallet.Address) (pwallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	onChainAcc := accounts.Account{Address: pethwallet.AsEthAddr(onChainAddr)}
	addr, err := pethchannel.DeployETHAssetholder(ctx, *cb.Cb, pethwallet.AsEthAddr(adjAddr), onChainAcc)
	return pethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying asset contract")
}
