package solana

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModule_CreateAccount(t *testing.T) {
	os.Setenv("SOLANA_USE_SANDBOX", "true")

	ctx := context.Background()

	module, err := New(ctx)
	require.NoError(t, err)

	account, err := module.CreateAccount(ctx, "token.json")
	require.NoError(t, err)
	require.NotNil(t, account)
}

func Test_loadFromKeyFile(t *testing.T) {
	account, err := loadFromKeyFile(context.Background(), "token.json")
	require.NoError(t, err)
	require.NotNil(t, account)
}
