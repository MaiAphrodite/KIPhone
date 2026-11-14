package main

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

func main() {
	w := solana.NewWallet()
	fmt.Println("pubkey  :", w.PublicKey())
	fmt.Println("keypair :", w.PrivateKey.String())
}
