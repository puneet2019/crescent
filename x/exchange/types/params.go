package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

var (
	KeyMarketCreationFee   = []byte("MarketCreationFee")
	KeyDefaultMakerFeeRate = []byte("DefaultMakerFeeRate")
	KeyDefaultTakerFeeRate = []byte("DefaultTakerFeeRate")
)

var (
	DefaultMarketCreationFee   = sdk.NewCoins()
	DefaultDefaultMakerFeeRate = sdk.NewDecWithPrec(-15, 4) // -0.15%
	DefaultDefaultTakerFeeRate = sdk.NewDecWithPrec(3, 3)   // 0.3%

	MinPrice = sdk.NewDecWithPrec(1, 14)
	MaxPrice = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 40))
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{
		MarketCreationFee:   DefaultMarketCreationFee,
		DefaultMakerFeeRate: DefaultDefaultMakerFeeRate,
		DefaultTakerFeeRate: DefaultDefaultTakerFeeRate,
	}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMarketCreationFee, &params.MarketCreationFee, validateMarketCreationFee),
		paramstypes.NewParamSetPair(KeyDefaultMakerFeeRate, &params.DefaultMakerFeeRate, validateDefaultMakerFeeRate),
		paramstypes.NewParamSetPair(KeyDefaultTakerFeeRate, &params.DefaultTakerFeeRate, validateDefaultTakerFeeRate),
	}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.MarketCreationFee, validateMarketCreationFee},
		{params.DefaultMakerFeeRate, validateDefaultMakerFeeRate},
		{params.DefaultTakerFeeRate, validateDefaultTakerFeeRate},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	if params.DefaultMakerFeeRate.IsNegative() && params.DefaultMakerFeeRate.Neg().GT(params.DefaultTakerFeeRate) {
		return fmt.Errorf("negative default maker fee rate must not be greater than default taker fee rate")
	}
	return nil
}

func validateMarketCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid market creation fee: %w", err)
	}
	return nil
}

func validateDefaultMakerFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.GT(utils.OneDec) {
		return fmt.Errorf("default maker fee rate must not be greater than 1.0: %s", v)
	}
	if v.LT(utils.OneDec.Neg()) {
		return fmt.Errorf("default maker fee rate must not be less than -1.0: %s", v)
	}
	return nil
}

func validateDefaultTakerFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.GT(utils.OneDec) {
		return fmt.Errorf("default taker fee rate must not be greater than 1.0: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("default taker fee rate must not be negative: %s", v)
	}
	return nil
}
