package keeper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type IndexerDistributionEvent struct {
	BlockHeight              int64         `json:"blockHeight"`
	BlockTimeUnixMicro       int64         `json:"blockTimeUnixMicro"`
	Action                   string        `json:"action"`
	Amount                   sdktypes.Coin `json:"amount"`
	DelegatorAddress         string        `json:"delegatorAddress"`
	ValidatorOperatorAddress string        `json:"validatorOperatorAddress"`
	WithdrawAddress          string        `json:"withdrawAddress"`
}

type IndexerWriter struct {
	output string
	file   *os.File
}

func NewIndexerWriter(homePath string) *IndexerWriter {
	// Resolve output path.
	output := filepath.Join(homePath, "indexer", "distribution.txt")
	// Create folder if doesn't exist.
	dir := filepath.Dir(output)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Open output file, creating if doesn't exist.
	file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("[INDEXER][distribution] Failed to open output file. output=%v, error=%w", output, err))
	}

	return &IndexerWriter{
		output: output,
		file:   file,
	}
}

// Write event to file.
func (iw *IndexerWriter) Write(ctx *sdktypes.Context, action string, amount sdktypes.Coin, delegatorAddress string, validatorOperatorAddress string, withdrawAddress string) {
	// If checking TX (simulating, not actually executing), do not index.
	if ctx.IsCheckTx() {
		return
	}

	encoder := json.NewEncoder(iw.file)

	// Export event.
	encoder.Encode(IndexerDistributionEvent{
		BlockHeight:              ctx.BlockHeight(),
		BlockTimeUnixMicro:       ctx.BlockTime().UnixMicro(),
		Action:                   action,
		Amount:                   amount,
		DelegatorAddress:         delegatorAddress,
		ValidatorOperatorAddress: validatorOperatorAddress,
		WithdrawAddress:          withdrawAddress,
	})

	ctx.Logger().Info("[INDEXER][distribution] Exported events", "blockHeight", ctx.BlockHeight(), "action", action, "amount", amount, "delegatorAddress", delegatorAddress, "validatorOperatorAddress", validatorOperatorAddress, "output", "withdrawAddress", withdrawAddress, iw.output)
}

// Close file.
func (iw *IndexerWriter) Close() {
	iw.file.Close()
}
