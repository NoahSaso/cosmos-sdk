package keeper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type IndexerBankEntity struct {
	ModuleName string        `json:"moduleName"`
	Address    string        `json:"address"`
	Balance    sdktypes.Coin `json:"balance"`
}

type IndexerBankEvent struct {
	BlockHeight        int64             `json:"blockHeight"`
	BlockTimeUnixMicro int64             `json:"blockTimeUnixMicro"`
	Action             string            `json:"action"`
	Coin               sdktypes.Coin     `json:"coin"`
	From               IndexerBankEntity `json:"from"`
	To                 IndexerBankEntity `json:"to"`
	NewSupply          sdktypes.Coin     `json:"newSupply"`
}

type IndexerWriter struct {
	output string
	file   *os.File
}

func NewIndexerWriter(homePath string) *IndexerWriter {
	// Resolve output path.
	output := filepath.Join(homePath, "indexer", "bank.txt")
	// Create folder if doesn't exist.
	dir := filepath.Dir(output)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Open output file, creating if doesn't exist.
	file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("[INDEXER][bank] Failed to open output file. output=%v, error=%w", output, err))
	}

	return &IndexerWriter{
		output: output,
		file:   file,
	}
}

// Write event to file.
func (iw *IndexerWriter) Write(ctx *sdktypes.Context, action string, coin sdktypes.Coin, from IndexerBankEntity, to IndexerBankEntity, newSupply sdktypes.Coin) {
	// If checking TX (simulating, not actually executing), do not index.
	if ctx.IsCheckTx() {
		return
	}

	encoder := json.NewEncoder(iw.file)

	// Export event.
	encoder.Encode(IndexerBankEvent{
		BlockHeight:        ctx.BlockHeight(),
		BlockTimeUnixMicro: ctx.BlockTime().UnixMicro(),
		Action:             action,
		Coin:               coin,
		From:               from,
		To:                 to,
		NewSupply:          newSupply,
	})

	ctx.Logger().Info("[INDEXER][bank] Exported events", "blockHeight", ctx.BlockHeight(), "action", action, "coin", coin, "from", from, "to", to, "newSupply", newSupply, "output", iw.output)
}

// Close file.
func (iw *IndexerWriter) Close() {
	iw.file.Close()
}
