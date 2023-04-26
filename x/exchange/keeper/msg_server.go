package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) CreateMarket(goCtx context.Context, msg *types.MsgCreateMarket) (*types.MsgCreateMarketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	market, err := k.Keeper.CreateMarket(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.BaseDenom, msg.QuoteDenom)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateMarketResponse{MarketId: market.Id}, nil
}

func (k msgServer) PlaceLimitOrder(goCtx context.Context, msg *types.MsgPlaceLimitOrder) (*types.MsgPlaceLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	order, execQty, execQuote, err := k.Keeper.PlaceLimitOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Price, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceLimitOrderResponse{
		Rested:           order.Id > 0,
		OrderId:          order.Id,
		ExecutedQuantity: execQty,
		ExecutedQuote:    execQuote,
	}, nil
}

func (k msgServer) PlaceMarketOrder(goCtx context.Context, msg *types.MsgPlaceMarketOrder) (*types.MsgPlaceMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	execQty, execQuote, err := k.Keeper.PlaceMarketOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceMarketOrderResponse{
		ExecutedQuantity: execQty,
		ExecutedQuote:    execQuote,
	}, nil
}

func (k msgServer) CancelOrder(goCtx context.Context, msg *types.MsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Keeper.CancelOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelOrderResponse{}, nil
}

func (k msgServer) SwapExactIn(goCtx context.Context, msg *types.MsgSwapExactIn) (*types.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	output, err := k.Keeper.SwapExactIn(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.Routes, msg.Input, msg.MinOutput, false)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapExactInResponse{
		Output: output,
	}, nil
}
