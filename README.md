# Solana token manager

Personal crypto token (meme-coin) creation instrument.
Written with Golang. Can be used to create Solana accounts and tokens.
Support adding Metaplex metadata program (smart contract).

How-to:
1. (Optional) Create a new Solana account(wallet) for token's owner.
   go run main.go create_account --output-key-file=owner_key.json

2. (Optional) Add or airdrop SOL (about 0.00001 SOL or 10000 Lamports) coins to new owner account.

3. Just create a token with go run main.go create_token --owner-key-file=owner_key.json --initial-supply=3000000 --name=example --symbol=exmpl --uri=https://example.com
