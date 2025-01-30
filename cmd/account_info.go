package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kirill-a-belov/solana_token_manager/internal/solana"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
	"github.com/spf13/cobra"
	"log"
)

func accountInfoCMD(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.accountInfoCMD")
	defer span.Done()

	var (
		ownerKeyFilename string
	)

	cmd := &cobra.Command{
		Use:   "account_info",
		Short: "Check Solana account info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx = context.Background()

			m, err := solana.New(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			info, err := m.SolanaAccountInfo(ctx, ownerKeyFilename)
			if err != nil {
				log.Fatalln(err)
			}

			res, err := json.MarshalIndent(info, "", "    ")
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(string(res))
		},
	}

	cmd.Flags().StringVar(&ownerKeyFilename, "owner-key-file", "", "Enter name for a file with owner account key details")
	cmd.MarkFlagRequired("owner-key-file")

	return cmd
}
