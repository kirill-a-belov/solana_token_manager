# Solana Token Manager

**Solana Token Manager** is a Go-based CLI tool for creating and managing personal Solana tokens, including meme-coins.  
It supports account creation, token minting, and integration with the **Metaplex Metadata program** for NFT-like metadata.

---

## Project Overview

This tool is designed for:
- **Solana account management**
- **Token creation and management**
- **Metadata program integration** for tokens

It enables developers and researchers to experiment with **Solana blockchain token operations** without manually handling complex transactions.

---

## Features

- Create Solana accounts (wallets)
- Mint new SPL tokens
- Assign Metaplex metadata to tokens
- Support for token airdrops for testing
- Written in **Golang**, lightweight and portable

---

## How to Use

1. (Optional) Create a new Solana account for the token owner:

```bash
go run main.go create_account --output-key-file=owner_key.json
```
(Optional) Add SOL to the account for testing (e.g., 0.00001 SOL / 10,000 Lamports).

Create a token:
```
go run main.go create_token \
--owner-key-file=owner_key.json \
--initial-supply=3000000 \
--name=ExampleToken \
--symbol=EXMPL \
--uri=https://example.com
```
## Status

Open-source and research-oriented.
Useful for blockchain development, Solana token experimentation, and NFT metadata integration research. Contributions and forks are welcome.

## Tags
#Solana #token #blockchain #Golang #Metaplex #NFT
