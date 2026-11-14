package main

import (
    "context"
    "fmt"
    "os"
    "time"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
)

func main() {
    pk := solana.MustPublicKeyFromBase58(os.Args[1])
    endpoints := []string{
        "https://rpc.ankr.com/solana_devnet",
        "https://api.devnet.solana.com",
        "https://devnet.genesysgo.net",
    }
    for _, e := range endpoints {
        ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
        c := rpc.New(e)
        sig, err := c.RequestAirdrop(ctx, pk, 1_000_000_000, rpc.CommitmentConfirmed)
        if err != nil {
            fmt.Printf("%-40s FAIL %v\n", e, err)
        } else {
            fmt.Printf("%-40s OK %s\n", e, sig)
        }
        cancel()
    }
}
