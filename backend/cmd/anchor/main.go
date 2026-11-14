package main

// Spike: can this machine anchor a batch hash on Solana devnet?
// Run: go run ./cmd/anchor <hash>
// Prints the transaction signature + explorer link, or the precise failure.

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/memo"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	hash := "sha256:demo"
	if len(os.Args) > 1 {
		hash = os.Args[1]
	}

	client := rpc.New(rpc.DevNet_RPC)
	wallet := solana.NewWallet()
	fmt.Println("wallet:", wallet.PublicKey())

	sig, err := client.RequestAirdrop(ctx, wallet.PublicKey(), 1_000_000_000, rpc.CommitmentConfirmed)
	if err != nil {
		fmt.Println("AIRDROP_FAILED:", err)
		os.Exit(2)
	}
	fmt.Println("airdrop tx:", sig)
	time.Sleep(4 * time.Second)

	inst := memo.NewMemoInstructionBuilder().SetMessage([]byte(hash)).SetSigner(wallet.PublicKey()).Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{inst},
		solana.Hash{}, // blockhash injected below
		solana.TransactionPayer(wallet.PublicKey()),
	)
	if err != nil {
		fmt.Println("BUILD_FAILED:", err)
		os.Exit(3)
	}
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		fmt.Println("BLOCKHASH_FAILED:", err)
		os.Exit(4)
	}
	tx.Message.RecentBlockhash = bh.Value.Blockhash
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(wallet.PublicKey()) {
			return &wallet.PrivateKey
		}
		return nil
	}); err != nil {
		fmt.Println("SIGN_FAILED:", err)
		os.Exit(5)
	}

	sent, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{PreflightCommitment: rpc.CommitmentConfirmed})
	if err != nil {
		fmt.Println("SEND_FAILED:", err)
		os.Exit(6)
	}
	fmt.Println("ok, signature:", sent)
	fmt.Println("explorer: https://explorer.solana.com/tx/" + sent.String() + "?cluster=devnet")
}
