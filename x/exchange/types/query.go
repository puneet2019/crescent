package types

func NewMarketResponse(market Market, marketState MarketState) MarketResponse {
	return MarketResponse{
		Id:            market.Id,
		BaseDenom:     market.BaseDenom,
		QuoteDenom:    market.QuoteDenom,
		EscrowAddress: market.EscrowAddress,
		LastPrice:     marketState.LastPrice,
	}
}
