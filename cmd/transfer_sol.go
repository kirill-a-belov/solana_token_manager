package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kirill-a-belov/solana_token_manager/internal/solana"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

func transferSOLCMD(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.transferSOLCMD")
	defer span.Done()

	var (
		ownerKeyFilename string
		amountLamports   uint64
		toAddress        string
	)

	cmd := &cobra.Command{
		Use:   "transfer_sol",
		Short: "Transfer SOL to account",
		Run: func(cmd *cobra.Command, args []string) {
			ctx = context.Background()

			m, err := solana.New(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Printf("Transfer %v SOL to %s ARE YOU SURE? (type \"yes\")\n", float64(amountLamports)/1000000000, toAddress)
			var check string
			fmt.Scanln(&check)
			if check != "yes" {
				fmt.Println("Exiting...")
				return
			}

			if err = m.TransferSOL(ctx, &solana.TransferSOLRequest{
				OwnerKeyFilename: ownerKeyFilename,
				AmountLamports:   amountLamports,
				TargetAddress:    toAddress,
			}); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVar(&ownerKeyFilename, "owner-key-file", "", "Enter name for a file with owner account key details")
	cmd.MarkFlagRequired("owner-key-file")

	cmd.Flags().Uint64Var(&amountLamports, "amount-lamports", 0, "Enter amount for an initial supply token")
	cmd.MarkFlagRequired("amount-lamports")

	cmd.Flags().StringVar(&toAddress, "to-address", "", "Token metadata name")
	cmd.MarkFlagRequired("to-address")

	return cmd
}
