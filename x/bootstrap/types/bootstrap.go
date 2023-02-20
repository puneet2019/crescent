package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v4/types"
)

func DeriveBootstrapPoolEscrowAddress(id uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BootstrapPoolEscrowAddress/%d", id)))
}

func DeriveBootstrapPoolFeeCollectorAddress(id uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BootstrapPoolFeeCollectorAddress/%d", id)))
}

func NewBootstrapPool(id uint64, baseCoinDenom, QuoteCoinDenom string, pairId uint64, minPrice, maxPrice *sdk.Dec, stages []Stage, proposer sdk.AccAddress, params Params) BootstrapPool {
	return BootstrapPool{
		Id:                    id,
		BaseCoinDenom:         baseCoinDenom,
		QuoteCoinDenom:        QuoteCoinDenom,
		PairId:                pairId,
		MinPrice:              minPrice,
		MaxPrice:              maxPrice,
		Stages:                stages,
		ProposerAddress:       proposer.String(),
		EscrowAddress:         DeriveBootstrapPoolEscrowAddress(id).String(),
		FeeCollectorAddress:   DeriveBootstrapPoolFeeCollectorAddress(id).String(),
		CreationFeeRate:       params.CreationFeeRate,
		ProtocolFeeRate:       params.ProtocolFeeRate,
		InitialTradingFeeRate: params.InitialTradingFeeRate,
		TradingFeeRate:        params.TradingFeeRate,
	}
}

func (m BootstrapPool) GetProposer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ProposerAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m BootstrapPool) GetEscrowAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.EscrowAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m BootstrapPool) GetFeeCollector() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.FeeCollectorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m BootstrapPool) IsActive() bool {
	// TODO: checking available stages
	return false
}

func (m BootstrapPool) CurrentStage() (id int32, stage Stage) {
	// TODO: checking available stages
	return 0, Stage{
		StartTime: time.Time{},
		EndTime:   time.Time{},
	}
}

func (m BootstrapPool) Validate() error {
	// validate addresses
	_, err := sdk.AccAddressFromBech32(m.ProposerAddress)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(m.EscrowAddress)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(m.FeeCollectorAddress)
	if err != nil {
		return err
	}

	// validate denoms
	if err := sdk.ValidateDenom(m.BaseCoinDenom); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(m.QuoteCoinDenom); err != nil {
		return err
	}

	// validate min/max prices
	if m.MinPrice != nil && !m.MinPrice.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "min price should be positive")
	}

	if m.MaxPrice != nil && !m.MaxPrice.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "max price should be positive")
	}

	if m.MinPrice != nil && m.MaxPrice != nil && m.MaxPrice.LTE(*m.MinPrice) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "max price should be greater than min price")
	}

	// validate pair id
	if m.PairId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pair id")
	}

	// validate stages
	numOfStages := uint32(len(m.Stages))
	if numOfStages < MinNumOfStages || numOfStages > MaxNumOfStages {
		return fmt.Errorf("num of stages %d must be greater than MinNumOfStages %s and lesser than MaxNumOfStages %s", numOfStages, MinNumOfStages, MaxNumOfStages)
	}
	for i, s := range m.Stages {
		duration := s.EndTime.Sub(s.StartTime)
		if duration < MinStageDuration || duration > MaxStageDuration {
			return fmt.Errorf("stage duration %d must be greater than MinStageDuration %s and lesser than MaxStageDuration %s", duration, MinStageDuration, MaxStageDuration)
		}
		if !s.EndTime.After(s.StartTime) {
			return fmt.Errorf("stage end time %s must be greater than start time %s", s.EndTime.Format(time.RFC3339), s.StartTime.Format(time.RFC3339))
		}
		for _, so := range m.Stages[i+1:] {
			if utils.DateRangesOverlap(s.StartTime, s.EndTime, so.StartTime, so.EndTime) {
				return fmt.Errorf("stage periods cannot be overlapped %s ~ %s with %s ~ %s", s.StartTime.Format(time.RFC3339), s.EndTime.Format(time.RFC3339), so.StartTime.Format(time.RFC3339), so.EndTime.Format(time.RFC3339))
			}
		}
	}

	// validate fee rates
	if m.CreationFeeRate.IsNegative() {
		return fmt.Errorf("creation fee rate must not be negative: %s", m.CreationFeeRate)
	}
	if m.ProtocolFeeRate.IsNegative() {
		return fmt.Errorf("protocol fee rate must not be negative: %s", m.ProtocolFeeRate)
	}
	if m.InitialTradingFeeRate.IsNegative() {
		return fmt.Errorf("initial trading fee rate must not be negative: %s", m.InitialTradingFeeRate)
	}
	if m.TradingFeeRate.IsNegative() {
		return fmt.Errorf("trading fee rate must not be negative: %s", m.TradingFeeRate)
	}

	return nil
}

func NewOrderForInitialOrder(io InitialOrder, id, poolId uint64, height int64, orderer string) Order {
	return Order{
		Id:                 id,
		BootstrapPoolId:    poolId,
		MsgHeight:          height,
		Orderer:            orderer,
		Direction:          io.Direction,
		OfferCoin:          io.OfferCoin,
		RemainingOfferCoin: io.OfferCoin,
		// TODO:
		//ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		ReceivedCoin:   sdk.Coin{},
		Price:          io.Price,
		Status:         OrderStatusNotExecuted,
		AssociateStage: io.AssociateStage,
	}
}

func NewOrderForLimitOrder(msg *MsgLimitOrder, id, poolId uint64, offerCoin sdk.Coin, price sdk.Dec, height int64) Order {
	return Order{
		Id:                 id,
		BootstrapPoolId:    poolId,
		MsgHeight:          height,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		// TODO:
		//ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		ReceivedCoin:   sdk.Coin{},
		Price:          price,
		Status:         OrderStatusNotExecuted,
		AssociateStage: 0,
	}
}

func (m Order) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

// Validate validates Order for genesis.
func (m Order) Validate() error {
	if m.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if m.BootstrapPoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if m.MsgHeight == 0 {
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(m.Orderer); err != nil {
		return fmt.Errorf("invalid orderer address %s: %w", m.Orderer, err)
	}
	if m.Direction != OrderDirectionBuy && m.Direction != OrderDirectionSell {
		return fmt.Errorf("invalid direction: %s", m.Direction)
	}
	if err := m.OfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid offer coin %s: %w", m.OfferCoin, err)
	}
	if m.OfferCoin.IsZero() {
		return fmt.Errorf("offer coin must not be 0")
	}
	if err := m.RemainingOfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid remaining offer coin %s: %w", m.RemainingOfferCoin, err)
	}
	if m.OfferCoin.Denom != m.RemainingOfferCoin.Denom {
		return fmt.Errorf("offer coin denom %s != remaining offer coin denom %s", m.OfferCoin.Denom, m.RemainingOfferCoin.Denom)
	}
	if err := m.ReceivedCoin.Validate(); err != nil {
		return fmt.Errorf("invalid received coin %s: %w", m.ReceivedCoin, err)
	}
	if !m.Price.IsPositive() {
		return fmt.Errorf("price must be positive: %s", m.Price)
	}

	// TODO: add other validations
	//if !m.Amount.IsPositive() {
	//	return fmt.Errorf("amount must be positive: %s", m.Amount)
	//}
	//if m.OpenAmount.IsNegative() {
	//	return fmt.Errorf("open amount must not be negative: %s", m.OpenAmount)
	//}
	//if m.BatchId == 0 {
	//	return fmt.Errorf("batch id must not be 0")
	//}
	//if m.ExpireAt.IsZero() {
	//	return fmt.Errorf("no expiration info")
	//}
	if !m.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return nil
}

// SetStatus sets the order's status.
// SetStatus is to easily find locations where the status is changed.
func (m *Order) SetStatus(status OrderStatus) {
	m.Status = status
}

func (m *Order) IsInitial() bool {
	if m.AssociateStage == 0 {
		return true
	}
	return false
}

// IsValid returns true if the OrderStatus is one of:
// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched,
// OrderStatusCompleted, OrderStatusCanceled, OrderStatusExpired.
func (x OrderStatus) IsValid() bool {
	switch x {
	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched,
		OrderStatusCompleted, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// IsMatchable returns true if the OrderStatus is one of:
// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched.
func (x OrderStatus) IsMatchable() bool {
	switch x {
	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched:
		return true
	default:
		return false
	}
}

// CanBeExpired has the same condition as IsMatchable.
func (x OrderStatus) CanBeExpired() bool {
	return x.IsMatchable()
}

//// CanBeCanceled returns true if the OrderStatus is one of:
//// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched.
//func (status OrderStatus) CanBeCanceled() bool {
//	switch status {
//	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched:
//		return true
//	default:
//		return false
//	}
//}

//// IsCanceledOrExpired returns true if the OrderStatus is one of:
//// OrderStatusCanceled, OrderStatusExpired.
//func (status OrderStatus) IsCanceledOrExpired() bool {
//	switch status {
//	case OrderStatusCanceled, OrderStatusExpired:
//		return true
//	default:
//		return false
//	}
//}
//
//// ShouldBeDeleted returns true if the OrderStatus is one of:
//// OrderStatusCompleted, OrderStatusCanceled, OrderStatusExpired.
//func (status OrderStatus) ShouldBeDeleted() bool {
//	return status == OrderStatusCompleted || status.IsCanceledOrExpired()
//}

// MustMarshaOrder returns the Order bytes.
// It throws panic if it fails.
func MustMarshaOrder(cdc codec.BinaryCodec, order Order) []byte {
	return cdc.MustMarshal(&order)
}

// UnmarshalOrder returns the Order from bytes.
func UnmarshalOrder(cdc codec.BinaryCodec, value []byte) (order Order, err error) {
	err = cdc.Unmarshal(value, &order)
	return order, err
}

// MustUnmarshalOrder returns the Order from bytes.
// It throws panic if it fails.
func MustUnmarshalOrder(cdc codec.BinaryCodec, value []byte) Order {
	msg, err := UnmarshalOrder(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}

// TODO: GetFeeCollector
// TODO: GetEscrowAddress
// TODO: GetProposer

//func GetAccAddress(address string) sdk.AccAddress {
//	if address == "" {
//		return nil
//	}
//	addr, err := sdk.AccAddressFromBech32(address)
//	if err != nil {
//		panic(err)
//	}
//	return addr
//}
//
//func (mm Bootstrap) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(mm.Address)
//}
//
//func (mm Bootstrap) Validate() error {
//	return ValidateBootstrap(mm.Address, mm.PairId)
//}
//
//func (i Incentive) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(i.Address)
//}
//
//func (i Incentive) Validate() error {
//	_, err := sdk.AccAddressFromBech32(i.Address)
//	if err != nil {
//		return err
//	}
//	return i.Claimable.Validate()
//}
//
//func ValidateBootstrap(address string, pairId uint64) error {
//	_, err := sdk.AccAddressFromBech32(address)
//	if err != nil {
//		return err
//	}
//
//	if pairId == uint64(0) {
//		return ErrInvalidPairId
//	}
//	return nil
//}
//
//func (mm BootstrapHandle) Validate() error {
//	return ValidateBootstrap(mm.Address, mm.PairId)
//}
//
//func (mm BootstrapHandle) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(mm.Address)
//}
//
//func (id IncentiveDistribution) Validate() error {
//	if err := ValidateBootstrap(id.Address, id.PairId); err != nil {
//		return err
//	}
//	if len(id.Amount) == 0 {
//		return fmt.Errorf("incentive distribution amount should be not empty")
//	}
//	return id.Amount.Validate()
//}
//
//func (id IncentiveDistribution) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(id.Address)
//}
//
//func UnmarshalBootstrap(cdc codec.BinaryCodec, value []byte) (mm Bootstrap, err error) {
//	err = cdc.Unmarshal(value, &mm)
//	return mm, err
//}
//
//func (idr DepositRecord) Validate() error {
//	if err := ValidateBootstrap(idr.Address, idr.PairId); err != nil {
//		return err
//	}
//	return idr.Amount.Validate()
//}
//
//func (idr DepositRecord) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(idr.Address)
//}
