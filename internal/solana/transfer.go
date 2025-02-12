package solana

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/program/token"
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
			time.Sleep(10 * time.Second)
		}
	}
	fmt.Println("")

	m.log.Info(ctx, "transaction has been confirmed", txSignature)

	return nil
}

type TransferSPLTokenRequest struct {
	OwnerKeyFilename string
	TargetAddress    string
	Amount           uint64
	TokenMint        string
}

func (m *Module) TransferSPLToken(ctx context.Context, req *TransferSPLTokenRequest) error {
	_, span := tracer.Start(ctx, "pkg.payment.TransferSPLToken")
	defer span.End()

	fromAccount, err := loadFromKeyFile(ctx, req.OwnerKeyFilename)
	if err != nil {
		return errors.Wrap(err, "failed to load sender account")
	}

	recipientPubKey := common.PublicKeyFromString(req.TargetAddress)
	tokenMintPubKey := common.PublicKeyFromString(req.TokenMint)

	fromTokenAccount, _, err := common.FindAssociatedTokenAddress(fromAccount.PublicKey, tokenMintPubKey)
	if err != nil {
		return errors.Wrap(err, "failed to find sender token account")
	}

	ataAccount, _, err := common.FindAssociatedTokenAddress(recipientPubKey, tokenMintPubKey)
	if err != nil {
		return errors.Wrap(err, "failed to find recipient token account")
	}

	instructionList := []types.Instruction{}

	accountInfo, err := m.solanaClient.GetAccountInfo(ctx, ataAccount.ToBase58())
	if err != nil || accountInfo.Owner.Bytes() == nil || len(accountInfo.Owner.Bytes()) == 0 || accountInfo.Owner != common.TokenProgramID {
		m.log.Info(ctx, "recipient ATA not found, creating new one...")

		ataInstruction := associated_token_account.Create(associated_token_account.CreateParam{
			Funder:                 fromAccount.PublicKey,
			Owner:                  recipientPubKey,
			Mint:                   tokenMintPubKey,
			AssociatedTokenAccount: ataAccount,
		})

		instructionList = append(instructionList, ataInstruction)

		m.log.Info(ctx, "created ATA for recipient", ataAccount.ToBase58())
	}

	recentBlockhash, err := m.solanaClient.GetLatestBlockhash(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get recent blockhash")
	}

	transferInstruction := token.Transfer(token.TransferParam{
		From:   fromTokenAccount,
		To:     ataAccount,
		Amount: req.Amount,
		Auth:   fromAccount.PublicKey,
	})
	instructionList = append(instructionList, transferInstruction)

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{*fromAccount},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        fromAccount.PublicKey,
			RecentBlockhash: recentBlockhash.Blockhash,
			Instructions:    instructionList,
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
			time.Sleep(10 * time.Second)
		}
	}

	m.log.Info(ctx, "transaction has been confirmed", txSignature)

	return nil
}
