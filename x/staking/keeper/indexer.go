package keeper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type IndexerDelegator struct {
	Address string        `json:"address"`
	Shares  sdktypes.Dec  `json:"shares"`
	Tokens  sdktypes.Coin `json:"tokens"`
}

type IndexerValidator struct {
	OperatorAddress string        `json:"operatorAddress"`
	TotalShares     sdktypes.Dec  `json:"totalShares"`
	TotalTokens     sdktypes.Coin `json:"totalTokens"`
}

type IndexerStakingEvent struct {
	BlockHeight        int64            `json:"blockHeight"`
	BlockTimeUnixMicro int64            `json:"blockTimeUnixMicro"`
	Action             string           `json:"action"`
	Amount             sdktypes.Coin    `json:"amount"`
	Shares             sdktypes.Dec     `json:"shares"`
	Delegator          IndexerDelegator `json:"delegator"`
	Validator          IndexerValidator `json:"validator"`
}

type IndexerWriter struct {
	output string
	file   *os.File
}

func NewIndexerWriter(homePath string) *IndexerWriter {
	// Resolve output path.
	output := filepath.Join(homePath, "indexer", "staking.txt")
	// Create folder if doesn't exist.
	dir := filepath.Dir(output)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Open output file, creating if doesn't exist.
	file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("[INDEXER][staking] Failed to open output file. output=%v, error=%w", output, err))
	}

	return &IndexerWriter{
		output: output,
		file:   file,
	}
}

// Write event to file.
func (iw *IndexerWriter) Write(ctx *sdktypes.Context, action string, amount sdktypes.Coin, shares sdktypes.Dec, delegator IndexerDelegator, validator IndexerValidator) {
	// If checking TX (simulating, not actually executing), do not index.
	if ctx.IsCheckTx() {
		return
	}

	encoder := json.NewEncoder(iw.file)

	// Export event.
	encoder.Encode(IndexerStakingEvent{
		BlockHeight:        ctx.BlockHeight(),
		BlockTimeUnixMicro: ctx.BlockTime().UnixMicro(),
		Action:             action,
		Amount:             amount,
		Shares:             shares,
		Delegator:          delegator,
		Validator:          validator,
	})

	ctx.Logger().Info("[INDEXER][staking] Exported events", "blockHeight", ctx.BlockHeight(), "action", action, "amount", amount, "shares", shares, "delegator", delegator, "validator", validator, "output", iw.output)
}

// Close file.
func (iw *IndexerWriter) Close() {
	iw.file.Close()
}
