package main

// Batch receipt + optional Solana devnet anchoring (PRD §5, stretch).
//
// The hash is always computed offline from the disbursement rows, so the demo
// works with no network. Anchoring is a separate, explicit action that only
// runs when a funded devnet keypair is configured — it never sits in the
// critical path.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/memo"
	"github.com/gagliardetto/solana-go/rpc"
)

type Batch struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Periode    string    `gorm:"size:10;index" json:"periode"`
	Hash       string    `gorm:"size:80" json:"hash"`
	RecordN    int       `json:"record_n"`
	Signature  string    `gorm:"size:120" json:"signature"`
	Cluster    string    `gorm:"size:20" json:"cluster"`
	AnchoredAt time.Time `json:"anchored_at"`
}

// canonicalBatchHash hashes a stable, sorted projection of the batch so the
// same data always produces the same receipt.
func canonicalBatchHash(disbs []Disbursement) (string, int) {
	sort.Slice(disbs, func(i, j int) bool { return disbs[i].ID < disbs[j].ID })
	type row struct {
		ID      uint   `json:"id"`
		NIM     string `json:"nim"`
		SK      string `json:"sk"`
		Periode string `json:"periode"`
		Nominal int64  `json:"nominal"`
		Cair    string `json:"cair"`
	}
	rows := make([]row, 0, len(disbs))
	for _, d := range disbs {
		rows = append(rows, row{ID: d.ID, SK: d.SKNomor, Periode: d.Periode, Nominal: d.Nominal, Cair: d.TanggalCair.UTC().Format(time.RFC3339)})
	}
	payload, _ := json.Marshal(rows)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), len(rows)
}

func (a *App) batchReceipt(w http.ResponseWriter, r *http.Request) {
	var disbs []Disbursement
	if err := a.db.Find(&disbs).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "load batch")
		return
	}
	if len(disbs) == 0 {
		writeErr(w, http.StatusBadRequest, "batch kosong — jalankan seed dulu")
		return
	}
	hash, n := canonicalBatchHash(disbs)

	var last Batch
	a.db.Order("id desc").First(&last)
	resp := map[string]any{
		"hash":        hash,
		"record_n":    n,
		"anchor_ready": os.Getenv("SOLANA_KEYPAIR") != "",
		"cluster":     "devnet",
	}
	if last.Hash == hash && last.Signature != "" {
		resp["signature"] = last.Signature
		resp["explorer"] = "https://explorer.solana.com/tx/" + last.Signature + "?cluster=devnet"
		resp["anchored_at"] = last.AnchoredAt
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *App) anchorBatch(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("SOLANA_KEYPAIR")
	if secret == "" {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":  "SOLANA_KEYPAIR belum diset",
			"detail": "butuh wallet devnet yang sudah dibiayai (faucet.solana.com). Hash batch tetap tersedia offline.",
		})
		return
	}
	wallet, err := solana.PrivateKeyFromBase58(secret)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "SOLANA_KEYPAIR tidak valid")
		return
	}

	var disbs []Disbursement
	if err := a.db.Find(&disbs).Error; err != nil || len(disbs) == 0 {
		writeErr(w, http.StatusBadRequest, "batch kosong")
		return
	}
	hash, n := canonicalBatchHash(disbs)

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	client := rpc.New(rpc.DevNet_RPC)

	inst := memo.NewMemoInstructionBuilder().SetMessage([]byte(hash)).SetSigner(wallet.PublicKey()).Build()
	tx, err := solana.NewTransaction([]solana.Instruction{inst}, solana.Hash{}, solana.TransactionPayer(wallet.PublicKey()))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "build tx: "+err.Error())
		return
	}
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "rpc blockhash: "+err.Error())
		return
	}
	tx.Message.RecentBlockhash = bh.Value.Blockhash
	if _, err := tx.Sign(func(k solana.PublicKey) *solana.PrivateKey {
		if k.Equals(wallet.PublicKey()) {
			return &wallet
		}
		return nil
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, "sign: "+err.Error())
		return
	}
	sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{PreflightCommitment: rpc.CommitmentConfirmed})
	if err != nil {
		writeErr(w, http.StatusBadGateway, "send: "+err.Error())
		return
	}

	batch := Batch{Periode: disbs[0].Periode, Hash: hash, RecordN: n, Signature: sig.String(), Cluster: "devnet", AnchoredAt: time.Now()}
	if err := a.db.Create(&batch).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "simpan batch")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"hash":     hash,
		"record_n": n,
		"signature": sig.String(),
		"explorer": "https://explorer.solana.com/tx/" + sig.String() + "?cluster=devnet",
	})
}
