// Copyright (c) The Tellor Authors.
// Licensed under the MIT License.

package ops

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	tellorCommon "github.com/tellor-io/TellorMiner/pkg/common"
	"github.com/tellor-io/TellorMiner/pkg/contracts/getter"
	"github.com/tellor-io/TellorMiner/pkg/contracts/tellor"
	"github.com/tellor-io/TellorMiner/pkg/rpc"
	"github.com/tellor-io/TellorMiner/pkg/util"
)

/**
 * This is the operational transfer component. Its purpose is to transfer tellor tokens
 */

func prepareTransfer(amt *big.Int, ctx context.Context) (*bind.TransactOpts, error) {
	instance := ctx.Value(tellorCommon.ContractsGetterContextKey).(*getter.TellorGetters)
	senderPubAddr := ctx.Value(tellorCommon.PublicAddress).(common.Address)

	balance, err := instance.BalanceOf(nil, senderPubAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	fmt.Println("My balance", util.FormatERC20Balance(balance))
	if balance.Cmp(amt) < 0 {
		return nil, fmt.Errorf("insufficient balance (%s TRB), requested %s TRB",
			util.FormatERC20Balance(balance),
			util.FormatERC20Balance(amt))
	}
	auth, err := PrepareEthTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare ethereum transaction: %s", err.Error())
	}
	return auth, nil
}

func Transfer(toAddress common.Address, amt *big.Int, ctx context.Context) error {
	auth, err := prepareTransfer(amt, ctx)
	if err != nil {
		return err
	}

	instance2 := ctx.Value(tellorCommon.ContractsTellorContextKey).(*tellor.Tellor)
	tx, err := instance2.Transfer(auth, toAddress, amt)
	if err != nil {
		return fmt.Errorf("contract failed: %s", err.Error())
	}
	fmt.Printf("Transferred %s to %s... with tx:\n%s\n", util.FormatERC20Balance(amt), toAddress.String()[:12], tx.Hash().Hex())
	return nil
}

func Approve(_spender common.Address, amt *big.Int, ctx context.Context) error {
	auth, err := prepareTransfer(amt, ctx)
	if err != nil {
		return err
	}

	instance2 := ctx.Value(tellorCommon.ContractsTellorContextKey).(*tellor.Tellor)
	tx, err := instance2.Approve(auth, _spender, amt)
	if err != nil {
		return err
	}
	fmt.Printf("Approved %s to %s... with tx:\n%s\n", util.FormatERC20Balance(amt), _spender.String()[:12], tx.Hash().Hex())
	return nil
}

func Balance(ctx context.Context, addr common.Address) error {
	client := ctx.Value(tellorCommon.ClientContextKey).(rpc.ETHClient)

	ethBalance, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return fmt.Errorf("problem getting balance: %+v", err)
	}

	instance := ctx.Value(tellorCommon.ContractsGetterContextKey).(*getter.TellorGetters)
	trbBalance, err := instance.BalanceOf(nil, addr)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Printf("%s\n", addr.String())
	fmt.Printf("%10s ETH\n", util.FormatERC20Balance(ethBalance))
	fmt.Printf("%10s TRB\n", util.FormatERC20Balance(trbBalance))
	return nil
}