package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Querier is used as Keeper will have duplicate methods if used directly,
// and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Querier) AllSpotMarkets(c context.Context, req *types.QueryAllSpotMarketsRequest) (*types.QueryAllSpotMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	marketStore := prefix.NewStore(store, types.SpotMarketKeyPrefix)
	var marketResps []types.SpotMarketResponse
	pageRes, err := query.Paginate(marketStore, req.Pagination, func(key, value []byte) error {
		var market types.SpotMarket
		k.cdc.MustUnmarshal(value, &market)
		marketResps = append(marketResps, k.MakeSpotMarketResponse(ctx, market))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllSpotMarketsResponse{
		Markets:    marketResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) SpotMarket(c context.Context, req *types.QuerySpotMarketRequest) (*types.QuerySpotMarketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	market, found := k.GetSpotMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "market not found")
	}
	return &types.QuerySpotMarketResponse{Market: k.MakeSpotMarketResponse(ctx, market)}, nil
}

func (k Querier) SpotOrder(c context.Context, req *types.QuerySpotOrderRequest) (*types.QuerySpotOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	order, found := k.GetSpotOrder(ctx, req.OrderId)
	if !found {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return &types.QuerySpotOrderResponse{Order: order}, nil
}

func (k Querier) BestSwapExactInRoutes(c context.Context, req *types.QueryBestSwapExactInRoutesRequest) (*types.QueryBestSwapExactInRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	allRoutes := k.FindAllRoutes(ctx, req.Input.Denom, req.MinOutput.Denom)
	if len(allRoutes) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "no possible routes")
	}
	var (
		bestOutput = utils.ZeroInt
		bestRoutes []uint64
	)
	for _, routes := range allRoutes {
		output, err := k.SwapExactIn(ctx, sdk.AccAddress{}, routes, req.Input, req.MinOutput, true)
		if err != nil && !errors.Is(err, types.ErrInsufficientOutput) { // sanity check
			panic(err)
		}
		if err == nil {
			if output.Amount.GT(bestOutput) {
				bestOutput = output.Amount
				bestRoutes = routes
			}
		}
	}
	if bestOutput.LT(req.MinOutput.Amount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "no possible routes") // TODO: use different error
	}
	return &types.QueryBestSwapExactInRoutesResponse{
		Routes: bestRoutes,
		Output: sdk.NewCoin(req.MinOutput.Denom, bestOutput),
	}, nil
}

func (k Querier) MakeSpotMarketResponse(ctx sdk.Context, market types.SpotMarket) types.SpotMarketResponse {
	marketState := k.MustGetSpotMarketState(ctx, market.Id)
	return types.NewSpotMarketResponse(market, marketState)
}
