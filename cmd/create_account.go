package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/kirill-a-belov/solana_token_manager/internal/solana"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

func createAccountCMD(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.createAccountCMD")
	defer span.Done()

	var outputKeyFilename string

	cmd := &cobra.Command{
		Use:   "create_account",
		Short: "Create Solana blockchain account",
		Run: func(cmd *cobra.Command, args []string) {
			ctx = context.Background()

			m, err := solana.New(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			if _, err := m.CreateAccount(ctx, outputKeyFilename); err != nil {
				log.Fatalln(err)
			}

		},
	}

	cmd.Flags().StringVar(&outputKeyFilename, "output-key-file", "", "Enter name for a new file with account key details")
	cmd.MarkFlagRequired("output-key-file")

	return cmd
}
