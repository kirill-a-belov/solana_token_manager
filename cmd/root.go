package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

func New(ctx context.Context) *cobra.Command {
	span, _ := tracer.Start(ctx, "cmd.New")
	defer span.Done()

	cmd := &cobra.Command{
		Short: "Solana token management CLI",
	}
	cmd.AddCommand(
		createAccountCMD(ctx),
		createTokenCMD(ctx),
		accountInfoCMD(ctx),
	)

	return cmd
}
