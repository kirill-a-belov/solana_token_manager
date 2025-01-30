package solana

import (
	"context"
	"encoding/json"
	"os"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/metaplex/token_metadata"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"

	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

type keyFileContent struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func (m *Module) CreateAccount(ctx context.Context, outputKeyFilename string) (*types.Account, error) {
	_, span := tracer.Start(ctx, "pkg.payment.CreateAccount")
	defer span.End()

	account := types.NewAccount()
	m.log.Info(ctx, "creating account",
		"public_key", account.PublicKey,
	)

	if m.config.UseSandbox {
		const defaultSandboxAirdropAmount = 2

		amount := uint64(defaultSandboxAirdropAmount * 1e9)
		txHash, err := m.solanaClient.RequestAirdrop(ctx, account.PublicKey.ToBase58(), amount)
		if err != nil {
			m.log.Error(ctx, "airdrop to account",
				"account_public_key", account.PublicKey.ToBase58(),
				"error", err,
			)
		} else {
			m.log.Info(ctx, "airdrop to account",
				"account_public_key", account.PublicKey.ToBase58(),
				"amount", amount,
				"tx_hash", txHash,
			)
		}
	}

	keyContent := keyFileContent{
		PublicKey:  account.PublicKey.ToBase58(),
		PrivateKey: base58.Encode(account.PrivateKey),
	}

	data, err := json.MarshalIndent(keyContent, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "marshal keys")
	}

	if err := os.WriteFile(outputKeyFilename, data, 0600); err != nil {
		return nil, errors.Wrap(err, "write output key file")
	}

	m.log.Info(ctx, "created account output keyfile",
		"output_key_filename", outputKeyFilename,
	)

	return &account, nil
}

func loadFromKeyFile(ctx context.Context, keyFilename string) (*types.Account, error) {
	_, span := tracer.Start(ctx, "pkg.payment.loadFromKeyFile")
	defer span.End()

	data, err := os.ReadFile(keyFilename)
	if err != nil {
		return nil, errors.Wrap(err, "read key file")
	}

	var keys map[string]string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, errors.Wrap(err, "unmarshal key file")
	}

	account, err := types.AccountFromBase58(keys["private_key"])
	if err != nil {
		return nil, errors.Wrap(err, "restoring account from keyfile")
	}

	return &account, nil
}

type SolanaAccountOwnedToken struct {
	PublicKey     string `json:"public_key"`
	MintPublicKey string `json:"mint_public_key"`
	Amount        uint64 `json:"amount"`

	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Uri    string `json:"uri"`
}
type SolanaAccountInfoResponse struct {
	PublicKey       string                     `json:"public_key"`
	Balance         uint64                     `json:"balance"`
	IsSystem        bool                       `json:"is_system"`
	IsSmartContract bool                       `json:"is_smart_contract"`
	RentEpoch       uint64                     `json:"rent_epoch"`
	Data            string                     `json:"data"`
	OwnedTokens     []*SolanaAccountOwnedToken `json:"owned_tokens"`
}

func (m *Module) SolanaAccountInfo(ctx context.Context, keyFilename string) (*SolanaAccountInfoResponse, error) {
	_, span := tracer.Start(ctx, "pkg.payment.SolanaAccountInfo")
	defer span.End()

	ownerAccount, err := loadFromKeyFile(ctx, keyFilename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load owner account")
	}

	accountInfo, err := m.solanaClient.GetAccountInfo(ctx, ownerAccount.PublicKey.ToBase58())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account info")
	}

	isSystem := accountInfo.Owner.String() == common.SystemProgramID.String()

	tokenAccountList, err := m.solanaClient.GetTokenAccountsByOwnerByProgram(ctx, ownerAccount.PublicKey.ToBase58(), common.TokenProgramID.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token accounts by owner public key")
	}

	tokenList := make([]*SolanaAccountOwnedToken, len(tokenAccountList))
	for i, tokenAccount := range tokenAccountList {
		metadataKey, _, err := common.FindProgramAddress([][]byte{[]byte("metadata"), common.MetaplexTokenMetaProgramID.Bytes(), tokenAccount.Mint.Bytes()}, common.MetaplexTokenMetaProgramID)
		if err != nil {
			return nil, errors.Wrap(err, "calculate metadata key")
		}

		metadataAccount, err := m.solanaClient.GetAccountInfo(ctx, metadataKey.ToBase58())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get account metadata")
		}

		tokenList[i] = &SolanaAccountOwnedToken{
			PublicKey:     tokenAccount.PublicKey.ToBase58(),
			MintPublicKey: tokenAccount.Mint.ToBase58(),
			Amount:        tokenAccount.Amount,
		}

		if len(metadataAccount.Data) > 0 {
			md, err := token_metadata.MetadataDeserialize(metadataAccount.Data)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode metadata")
			}
			tokenList[i].Name = md.Data.Name
			tokenList[i].Symbol = md.Data.Symbol
			tokenList[i].Uri = md.Data.Uri

		}
	}

	res := &SolanaAccountInfoResponse{
		PublicKey:       ownerAccount.PublicKey.ToBase58(),
		Balance:         accountInfo.Lamports,
		IsSystem:        isSystem,
		IsSmartContract: accountInfo.Executable,
		RentEpoch:       accountInfo.RentEpoch,
		Data:            string(accountInfo.Data),
		OwnedTokens:     tokenList,
	}

	return res, nil
}
