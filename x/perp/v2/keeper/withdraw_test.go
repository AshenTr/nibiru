package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	. "github.com/NibiruChain/nibiru/x/common/testutil/action"
	. "github.com/NibiruChain/nibiru/x/common/testutil/assertion"
	types "github.com/NibiruChain/nibiru/x/perp/v1/types"
	. "github.com/NibiruChain/nibiru/x/perp/v2/integration/action"
	. "github.com/NibiruChain/nibiru/x/perp/v2/integration/assertion"
	v2types "github.com/NibiruChain/nibiru/x/perp/v2/types"
)

func TestWithdraw(t *testing.T) {
	alice := testutil.AccAddress()
	pairBtcUsdc := asset.Registry.Pair(denoms.BTC, denoms.USDC)
	startBlockTime := time.Now()

	tc := TestCases{
		TC("successful withdraw, no bad debt").
			Given(
				SetBlockNumber(1),
				SetBlockTime(startBlockTime),
				CreateCustomMarket(pairBtcUsdc),
				FundModule(v2types.VaultModuleAccount, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1000)))),
			).
			When(
				Withdraw(pairBtcUsdc, alice, sdk.NewInt(1000)),
			).
			Then(
				BalanceEqual(alice, denoms.USDC, sdk.NewInt(1000)),
				ModuleBalanceEqual(types.VaultModuleAccount, denoms.USDC, sdk.ZeroInt()),
				MarketShouldBeEqual(pairBtcUsdc, Market_PrepaidBadDebtShouldBeEqualTo(sdk.ZeroInt())),
			),

		TC("successful withdraw, some bad debt").
			Given(
				SetBlockNumber(1),
				SetBlockTime(startBlockTime),
				CreateCustomMarket(pairBtcUsdc),
				FundModule(v2types.VaultModuleAccount, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(500)))),
				FundModule(v2types.PerpEFModuleAccount, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(500)))),
			).
			When(
				Withdraw(pairBtcUsdc, alice, sdk.NewInt(1000)),
			).
			Then(
				BalanceEqual(alice, denoms.USDC, sdk.NewInt(1000)),
				ModuleBalanceEqual(types.VaultModuleAccount, denoms.USDC, sdk.ZeroInt()),
				ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.ZeroInt()),
				MarketShouldBeEqual(pairBtcUsdc, Market_PrepaidBadDebtShouldBeEqualTo(sdk.NewInt(500))),
			),

		TC("successful withdraw, all bad debt").
			Given(
				SetBlockNumber(1),
				SetBlockTime(startBlockTime),
				CreateCustomMarket(pairBtcUsdc),
				FundModule(v2types.PerpEFModuleAccount, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1000)))),
			).
			When(
				Withdraw(pairBtcUsdc, alice, sdk.NewInt(1000)),
			).
			Then(
				BalanceEqual(alice, denoms.USDC, sdk.NewInt(1000)),
				ModuleBalanceEqual(types.VaultModuleAccount, denoms.USDC, sdk.ZeroInt()),
				ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.ZeroInt()),
				MarketShouldBeEqual(pairBtcUsdc, Market_PrepaidBadDebtShouldBeEqualTo(sdk.NewInt(1000))),
			),

		TC("successful withdraw, existing bad debt").
			Given(
				SetBlockNumber(1),
				SetBlockTime(startBlockTime),
				CreateCustomMarket(pairBtcUsdc, WithPrepaidBadDebt(sdk.NewInt(1000))),
				FundModule(v2types.PerpEFModuleAccount, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1000)))),
			).
			When(
				Withdraw(pairBtcUsdc, alice, sdk.NewInt(1000)),
			).
			Then(
				BalanceEqual(alice, denoms.USDC, sdk.NewInt(1000)),
				ModuleBalanceEqual(types.VaultModuleAccount, denoms.USDC, sdk.ZeroInt()),
				ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.ZeroInt()),
				MarketShouldBeEqual(pairBtcUsdc, Market_PrepaidBadDebtShouldBeEqualTo(sdk.NewInt(2000))),
			),
	}

	NewTestSuite(t).WithTestCases(tc...).Run()
}
