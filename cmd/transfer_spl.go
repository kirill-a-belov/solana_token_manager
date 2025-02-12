package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kirill-a-belov/solana_token_manager/internal/solana"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

func transferSPLCMD(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.transferSPLCMD")
	defer span.Done()

	var (
		ownerKeyFilename string
		amountTokens     uint64
		toAddress        string
		tokenMint        string
	)

	cmd := &cobra.Command{
		Use:   "transfer_spl",
		Short: "Transfer SOL token to account",
		Run: func(cmd *cobra.Command, args []string) {
			ctx = context.Background()

			m, err := solana.New(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Printf("Transfer %v SPL to %s ARE YOU SURE? (type \"yes\")\n", amountTokens, toAddress)
			var check string
			fmt.Scanln(&check)
			if check != "yes" {
				fmt.Println("Exiting...")
				return
			}

			if err = m.TransferSPLToken(ctx, &solana.TransferSPLTokenRequest{
				OwnerKeyFilename: ownerKeyFilename,
				TargetAddress:    toAddress,
				Amount:           amountTokens,
				TokenMint:        tokenMint,
			}); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVar(&ownerKeyFilename, "owner-key-file", "", "Enter name for a file with owner account key details")
	cmd.MarkFlagRequired("owner-key-file")

	cmd.Flags().Uint64Var(&amountTokens, "amount-tokens", 0, "Enter amount for an initial supply token")
	cmd.MarkFlagRequired("amount-tokens")

	cmd.Flags().StringVar(&toAddress, "to-address", "", "Recipient address")
	cmd.MarkFlagRequired("to-address")

	cmd.Flags().StringVar(&tokenMint, "token-mint", "", "Token mint")
	cmd.MarkFlagRequired("to-address")

	return cmd
}
