package keeper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type IndexerSlashEvent struct {
	Type                         string       `json:"type"`
	RegisteredBlockHeight        int64        `json:"registeredBlockHeight"`
	RegisteredBlockTimeUnixMicro int64        `json:"registeredBlockTimeUnixMicro"`
	InfractionBlockHeight        int64        `json:"infractionBlockHeight"`
	ValidatorOperator            string       `json:"validatorOperator"`
	SlashFactor                  sdktypes.Dec `json:"slashFactor"`
	AmountSlashed                sdktypes.Int `json:"amountSlashed"`
}

type IndexerWriter struct {
	output string
	file   *os.File
}

func NewIndexerWriter(homePath string) *IndexerWriter {
	// Resolve output path.
	output := filepath.Join(homePath, "indexer", "staking.out")
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

// Write slash event to file.
func (iw *IndexerWriter) WriteSlash(ctx *sdktypes.Context, infractionBlockHeight int64, validatorOperator string, slashFactor sdktypes.Dec, amountSlashed sdktypes.Int) {
	// If checking TX (simulating, not actually executing), do not index.
	if ctx.IsCheckTx() {
		return
	}

	encoder := json.NewEncoder(iw.file)

	// Export event.
	event := IndexerSlashEvent{
		Type:                         "slash",
		RegisteredBlockHeight:        ctx.BlockHeight(),
		RegisteredBlockTimeUnixMicro: ctx.BlockTime().UnixMicro(),
		InfractionBlockHeight:        infractionBlockHeight,
		ValidatorOperator:            validatorOperator,
		SlashFactor:                  slashFactor,
		AmountSlashed:                amountSlashed,
	}
	encoder.Encode(event)

	ctx.Logger().Info("[INDEXER][staking] Exported event", "type", event.Type, "registeredBlockHeight", event.RegisteredBlockHeight, "registeredBlockTimeUnixMicro", event.RegisteredBlockTimeUnixMicro, "infractionBlockHeight", event.InfractionBlockHeight, "validatorOperator", event.ValidatorOperator, "slashFactor", event.SlashFactor, "amountSlashed", event.AmountSlashed, "output", iw.output)
}

// Close file.
func (iw *IndexerWriter) Close() {
	iw.file.Close()
}
