package keeper_test

import (
	"testing"

	"github.com/NibiruChain/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	"github.com/NibiruChain/nibiru/x/common/testutil/testapp"
	perpammtypes "github.com/NibiruChain/nibiru/x/perp/v1/amm/types"
	"github.com/NibiruChain/nibiru/x/perp/v1/keeper"
	"github.com/NibiruChain/nibiru/x/perp/v1/types"
)

func TestFrom2To3(t *testing.T) {
	alice := testutil.AccAddress()
	bob := testutil.AccAddress()
	pairBtcUsd := asset.Registry.Pair(denoms.BTC, denoms.NUSD)

	testCases := []struct {
		name               string
		positions          []types.Position
		expectedTotalLong  sdk.Dec
		expectedTotalShort sdk.Dec
	}{
		{
			"one position long",
			[]types.Position{
				{
					TraderAddress: alice.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
			},
			sdk.MustNewDecFromStr("10000000"),
			sdk.ZeroDec(),
		},
		{
			"two long positions",
			[]types.Position{
				{
					TraderAddress: alice.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
				{
					TraderAddress: bob.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
			},
			sdk.MustNewDecFromStr("20000000"),
			sdk.ZeroDec(),
		},
		{
			"one long position and one short position: long bigger than short",
			[]types.Position{
				{
					TraderAddress: alice.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
				{
					TraderAddress: bob.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("-5000000"),
				},
			},
			sdk.MustNewDecFromStr("10000000"),
			sdk.MustNewDecFromStr("5000000"),
		},
		{
			"one long position and one short position: short bigger than long",
			[]types.Position{
				{
					TraderAddress: alice.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
				{
					TraderAddress: bob.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("-25000000"),
				},
			},
			sdk.MustNewDecFromStr("10000000"),
			sdk.MustNewDecFromStr("25000000"),
		},
		{
			"one long position and one short position: equal long and short",
			[]types.Position{
				{
					TraderAddress: alice.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("10000000"),
				},
				{
					TraderAddress: bob.String(),
					Pair:          pairBtcUsd,
					Size_:         sdk.MustNewDecFromStr("-10000000"),
				},
			},
			sdk.MustNewDecFromStr("10000000"),
			sdk.MustNewDecFromStr("10000000"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			app, ctx := testapp.NewNibiruTestAppAndContext(true)

			market := perpammtypes.Market{
				Pair:         asset.Registry.Pair(denoms.BTC, denoms.NUSD),
				BaseReserve:  sdk.MustNewDecFromStr("10000000"),
				QuoteReserve: sdk.MustNewDecFromStr("20000000"),
				TotalLong:    sdk.ZeroDec(),
				TotalShort:   sdk.ZeroDec(),
			}
			app.PerpAmmKeeper.Pools.Insert(ctx, market.Pair, market)

			savedPool, err := app.PerpAmmKeeper.Pools.Get(ctx, market.Pair)
			require.NoError(t, err)
			require.Equal(t, sdk.ZeroDec(), savedPool.TotalLong)
			require.Equal(t, sdk.ZeroDec(), savedPool.TotalShort)
			require.Equal(t, sdk.ZeroDec(), savedPool.PegMultiplier)

			for _, pos := range tc.positions {
				addr, err := sdk.AccAddressFromBech32(pos.TraderAddress)
				if err != nil {
					t.Errorf("invalid address: %s", pos.TraderAddress)
				}
				app.PerpKeeper.Positions.Insert(ctx, collections.Join(pos.Pair, addr), pos)
			}

			// Run migration
			err = keeper.From2To3(app.PerpKeeper, app.PerpAmmKeeper)(ctx)
			require.NoError(t, err)

			savedPool, err = app.PerpAmmKeeper.Pools.Get(ctx, market.Pair)
			require.NoError(t, err)
			require.Equal(t, tc.expectedTotalLong, savedPool.TotalLong)
			require.Equal(t, tc.expectedTotalShort, savedPool.TotalShort)
			require.Equal(t, sdk.OneDec(), savedPool.PegMultiplier)
		})
	}
}
