package solana

import (
	"context"
	"fmt"
	"time"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/pkg/errors"

	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

type TransferSOLRequest struct {
	OwnerKeyFilename string
	TargetAddress    string
	AmountLamports   uint64
}

func (m *Module) TransferSOL(ctx context.Context, req *TransferSOLRequest) error {
	_, span := tracer.Start(ctx, "pkg.payment.TransferSOL")
	defer span.End()

	fromAccount, err := loadFromKeyFile(ctx, req.OwnerKeyFilename)
	if err != nil {
		return errors.Wrap(err, "failed to load sender account")
	}

	recipientPubKey := common.PublicKeyFromString(req.TargetAddress)

	ownerBalance, err := m.solanaClient.GetBalance(ctx, fromAccount.PublicKey.ToBase58())
	if err != nil {
		return errors.Wrap(err, "failed to get sender balance")
	}
	m.log.Info(ctx, "sender balance", ownerBalance)

	recentBlockhash, err := m.solanaClient.GetLatestBlockhash(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get recent blockhash")
	}
	m.log.Info(ctx, "recent Solana block hash", recentBlockhash.Blockhash)

	feeCalculator, err := m.solanaClient.GetFeeForMessage(ctx, types.NewMessage(types.NewMessageParam{
		FeePayer:        fromAccount.PublicKey,
		RecentBlockhash: recentBlockhash.Blockhash,
		Instructions:    []types.Instruction{},
	}))
	if err != nil {
		return errors.Wrap(err, "failed to get transaction fee")
	}
	m.log.Info(ctx, "estimated transaction fee", *feeCalculator)

	totalAmount := req.AmountLamports + *feeCalculator
	if ownerBalance < totalAmount {
		return errors.New("insufficient balance for transfer and fees")
	}

	transferInstruction := system.Transfer(system.TransferParam{
		From:   fromAccount.PublicKey,
		To:     recipientPubKey,
		Amount: req.AmountLamports,
	})

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{*fromAccount},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        fromAccount.PublicKey,
			RecentBlockhash: recentBlockhash.Blockhash,
			Instructions:    []types.Instruction{transferInstruction},
		}),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}

	txSignature, err := m.solanaClient.SendTransaction(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "failed to send transaction")
	}
	m.log.Info(ctx, "transaction signature", txSignature)

	// Wait for confirmation
	txNotConfirmed := true
	fmt.Print("waiting for confirmation")
	for txNotConfirmed {
		tx, err := m.solanaClient.GetTransaction(ctx, txSignature)
		if err != nil {
			m.log.Error(ctx, "failed to get transaction", err, txSignature)
		}
		if tx != nil && tx.Meta != nil && tx.Meta.Err == nil {
			txNotConfirmed = false
		} else {
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Println("")

	m.log.Info(ctx, "transaction has been confirmed", txSignature)

	return nil
}
