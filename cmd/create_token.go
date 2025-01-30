package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/kirill-a-belov/solana_token_manager/internal/solana"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

func createTokenCMD(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.createTokenCMD")
	defer span.Done()

	var (
		outputKeyFilename string
		ownerKeyFilename  string
		initialSupply     uint64

		name   string
		symbol string
		uri    string
	)

	cmd := &cobra.Command{
		Use:   "create_token",
		Short: "Create Solana token",
		Run: func(cmd *cobra.Command, args []string) {
			ctx = context.Background()

			m, err := solana.New(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			if err = m.CreateToken(ctx, &solana.CreateTokenRequest{
				OutputTokenKeyFilename: outputKeyFilename,
				OwnerKeyFilename:       ownerKeyFilename,
				InitialSupply:          initialSupply,
				Name:                   name,
				Symbol:                 symbol,
				Uri:                    uri,
			}); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVar(&ownerKeyFilename, "owner-key-file", "", "Enter name for a file with owner account key details")
	cmd.MarkFlagRequired("owner-key-file")

	cmd.Flags().StringVar(&outputKeyFilename, "output-key-file", "new_token_key.json", "Enter name for a new file with token key details")
	cmd.Flags().Uint64Var(&initialSupply, "initial-supply", 0, "Enter amount for an initial supply token")

	cmd.Flags().StringVar(&name, "name", "", "Token metadata name")
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringVar(&symbol, "symbol", "", "Token metadata symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringVar(&uri, "uri", "", "Token metadata Uri")
	cmd.MarkFlagRequired("uri")

	return cmd
}
