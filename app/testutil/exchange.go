package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *TestSuite) CreateMarket(creatorAddr sdk.AccAddress, baseDenom, quoteDenom string, fundFee bool) exchangetypes.Market {
	s.T().Helper()
	if fundFee {
		s.FundAccount(creatorAddr, s.App.ExchangeKeeper.GetMarketCreationFee(s.Ctx))
	}
	market, err := s.App.ExchangeKeeper.CreateMarket(s.Ctx, creatorAddr, baseDenom, quoteDenom)
	s.Require().NoError(err)
	return market
}

func (s *TestSuite) PlaceLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price sdk.Dec, qty sdk.Int) (order exchangetypes.Order, execQty, execQuote sdk.Int) {
	s.T().Helper()
	var err error
	order, execQty, execQuote, err = s.App.ExchangeKeeper.PlaceLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceMarketOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, qty sdk.Int) (execQty, execQuote sdk.Int) {
	s.T().Helper()
	var err error
	execQty, execQuote, err = s.App.ExchangeKeeper.PlaceMarketOrder(s.Ctx, marketId, ordererAddr, isBuy, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) SwapExactIn(
	ordererAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.Coin, simulate bool) (output sdk.Coin) {
	s.T().Helper()
	var err error
	output, err = s.App.ExchangeKeeper.SwapExactIn(s.Ctx, ordererAddr, routes, input, minOutput, simulate)
	s.Require().NoError(err)
	return output
}
