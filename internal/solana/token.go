package solana

import (
	"context"
	"fmt"
	"time"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/program/metaplex/token_metadata"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/pkg/errors"

	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

type CreateTokenRequest struct {
	OwnerKeyFilename       string
	OutputTokenKeyFilename string
	InitialSupply          uint64

	Name   string
	Symbol string
	Uri    string
}

func (m *Module) CreateToken(ctx context.Context, req *CreateTokenRequest) error {
	_, span := tracer.Start(ctx, "pkg.payment.CreateToken")
	defer span.End()

	ownerAccount, err := loadFromKeyFile(ctx, req.OwnerKeyFilename)
	if err != nil {
		return errors.Wrap(err, "failed to load owner account")
	}

	ownerBalance, err := m.solanaClient.GetBalance(ctx, ownerAccount.PublicKey.ToBase58())
	if err != nil {
		return errors.Wrap(err, "get owner balance")
	}
	m.log.Info(ctx, "owner balance", ownerBalance)

	mintAccount, err := m.CreateAccount(ctx, req.OutputTokenKeyFilename)
	if err != nil {
		return errors.Wrap(err, "create mint account")
	}
	m.log.Info(ctx, "mint account address", mintAccount.PublicKey.ToBase58())

	exemptionMinBalance, err := m.solanaClient.GetMinimumBalanceForRentExemption(ctx, token.MintAccountSize)
	if err != nil {
		return errors.Wrap(err, "get exemption min balance")
	}
	m.log.Info(ctx, "exemption min balance", exemptionMinBalance)

	createMintAccountInstruction := system.CreateAccount(system.CreateAccountParam{
		From:     ownerAccount.PublicKey,
		New:      mintAccount.PublicKey,
		Lamports: exemptionMinBalance,
		Space:    token.MintAccountSize,
		Owner:    common.TokenProgramID,
	})

	initializeMintInstruction := token.InitializeMint(token.InitializeMintParam{
		Decimals: 0,
		Mint:     mintAccount.PublicKey,
		MintAuth: ownerAccount.PublicKey,
	})

	ataAddress, _, err := common.FindAssociatedTokenAddress(ownerAccount.PublicKey, mintAccount.PublicKey)
	if err != nil {
		return errors.Wrap(err, "calculate ATA address")
	}
	m.log.Info(ctx, "ATA account address", ataAddress.ToBase58())

	createATAInstruction := associated_token_account.Create(associated_token_account.CreateParam{
		Funder:                 ownerAccount.PublicKey,
		Owner:                  ownerAccount.PublicKey,
		Mint:                   mintAccount.PublicKey,
		AssociatedTokenAccount: ataAddress,
	})

	mintToInstruction := token.MintTo(token.MintToParam{
		Mint:   mintAccount.PublicKey,
		Auth:   ownerAccount.PublicKey,
		To:     ataAddress,
		Amount: req.InitialSupply,
	})

	mintData := token_metadata.DataV2{
		Name:                 req.Name,
		Symbol:               req.Symbol,
		Uri:                  req.Uri,
		SellerFeeBasisPoints: 0,
		Creators: &[]token_metadata.Creator{
			{
				Address:  ownerAccount.PublicKey,
				Verified: true,
				Share:    100,
			},
		},
	}

	metadataKey, _, err := common.FindProgramAddress([][]byte{[]byte("metadata"), common.MetaplexTokenMetaProgramID.Bytes(), mintAccount.PublicKey.Bytes()}, common.MetaplexTokenMetaProgramID)
	if err != nil {
		return errors.Wrap(err, "calculate metadata key")
	}

	metadataInstruction := token_metadata.CreateMetadataAccountV3(token_metadata.CreateMetadataAccountV3Param{
		Metadata:                metadataKey,
		Mint:                    mintAccount.PublicKey,
		MintAuthority:           ownerAccount.PublicKey,
		Payer:                   ownerAccount.PublicKey,
		UpdateAuthority:         ownerAccount.PublicKey,
		UpdateAuthorityIsSigner: true,
		IsMutable:               true,
		Data:                    mintData,
	})

	recentBlockhash, err := m.solanaClient.GetLatestBlockhash(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get recent blockhash")
	}
	m.log.Info(ctx, "recent Solana block hash", recentBlockhash)

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{*ownerAccount, *mintAccount},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        ownerAccount.PublicKey,
			RecentBlockhash: recentBlockhash.Blockhash,
			Instructions: []types.Instruction{
				createMintAccountInstruction,
				initializeMintInstruction,
				createATAInstruction,
				mintToInstruction,
				metadataInstruction,
			},
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
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Println("")

	m.log.Info(ctx, "transaction has been confirmed", txSignature)

	return nil
}
