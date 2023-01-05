package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	totalSupply := sdk.Coins{}

	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)
	for _, balance := range genState.Balances {
		addr := balance.GetAddress()

		if err := k.initBalances(ctx, addr, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if !genState.Supply.Empty() && !genState.Supply.IsEqual(totalSupply) {
		panic(fmt.Errorf("genesis supply is incorrect, expected %v, got %v", genState.Supply, totalSupply))
	}

	// INDEXER.
	for _, balance := range genState.Balances {
		for _, coin := range balance.Coins {
			k.indexerWriter.Write(
				&ctx,
				"genesis_balance",
				coin,
				// Empty "from" for initial balance.
				IndexerBankEntity{
					ModuleName: "",
					Address:    "",
					Balance: sdk.Coin{
						Denom:  coin.GetDenom(),
						Amount: sdk.NewInt(-1),
					},
				},
				IndexerBankEntity{
					ModuleName: "",
					Address:    balance.Address,
					Balance:    k.GetBalance(ctx, balance.GetAddress(), coin.GetDenom()),
				},
				// Genesis supply indexed below, not here.
				sdk.Coin{
					Denom:  coin.GetDenom(),
					Amount: sdk.NewInt(-1),
				},
			)
		}
	}

	for _, supply := range totalSupply {
		k.setSupply(ctx, supply)

		// INDEXER.
		k.indexerWriter.Write(
			&ctx,
			"genesis_supply",
			supply,
			// Empty "from" and "to" for initial supply.
			IndexerBankEntity{
				ModuleName: "",
				Address:    "",
				Balance: sdk.Coin{
					Denom:  supply.GetDenom(),
					Amount: sdk.NewInt(-1),
				},
			},
			IndexerBankEntity{
				ModuleName: "",
				Address:    "",
				Balance: sdk.Coin{
					Denom:  supply.GetDenom(),
					Amount: sdk.NewInt(-1),
				},
			},
			supply,
		)
	}

	for _, meta := range genState.DenomMetadata {
		k.SetDenomMetaData(ctx, meta)
	}
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	totalSupply, _, err := k.GetPaginatedTotalSupply(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(fmt.Errorf("unable to fetch total supply %v", err))
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAccountsBalances(ctx),
		totalSupply,
		k.GetAllDenomMetaData(ctx),
	)
}
