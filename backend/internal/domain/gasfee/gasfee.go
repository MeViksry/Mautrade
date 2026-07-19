package gasfee

import (
	"fmt"

	"github.com/MeViksry/qdecimal"
)

type Type string

const (
	TypeProfitShare Type = "profit_share"
	TypeLossRebate  Type = "loss_rebate"
	TypeBreakeven   Type = "breakeven"
)

type Result struct {
	Type            Type             `json:"type"`
	GrossPnL        qdecimal.Decimal `json:"gross_pnl"`
	GasFeeAmount    qdecimal.Decimal `json:"gas_fee_amount"`
	PlatformRebate  qdecimal.Decimal `json:"platform_rebate"`
	NetAmountToUser qdecimal.Decimal `json:"net_amount_to_user"`
	NetLossToUser   qdecimal.Decimal `json:"net_loss_to_user"`
}

type Calculator struct {
	ShareRate qdecimal.Decimal
}

func NewCalculator(shareRate qdecimal.Decimal) (Calculator, error) {
	if shareRate.Sign() < 0 || shareRate.Cmp(qdecimal.One) > 0 {
		return Calculator{}, fmt.Errorf("gasfee: share rate must be between 0 and 1")
	}
	return Calculator{ShareRate: shareRate}, nil
}

func MustCalculator(shareRate string) Calculator {
	rate := qdecimal.MustParse(shareRate)
	calculator, err := NewCalculator(rate)
	if err != nil {
		panic(err)
	}
	return calculator
}

func (c Calculator) CalculateFromPrices(entryPrice, exitPrice, quantity qdecimal.Decimal) Result {
	entryValue := entryPrice.Mul(quantity)
	exitValue := exitPrice.Mul(quantity)
	return c.CalculateFromValues(entryValue, exitValue)
}

func (c Calculator) CalculateFromValues(entryValue, exitValue qdecimal.Decimal) Result {
	grossPnL := exitValue.Sub(entryValue)
	zero := qdecimal.Zero

	switch grossPnL.Sign() {
	case 1:
		gasFee := grossPnL.Mul(c.ShareRate)
		return Result{
			Type:            TypeProfitShare,
			GrossPnL:        grossPnL,
			GasFeeAmount:    gasFee,
			NetAmountToUser: grossPnL.Sub(gasFee),
			PlatformRebate:  zero,
			NetLossToUser:   zero,
		}
	case -1:
		loss := grossPnL.Abs()
		rebate := loss.Mul(c.ShareRate)
		return Result{
			Type:            TypeLossRebate,
			GrossPnL:        grossPnL,
			GasFeeAmount:    rebate.Neg(),
			PlatformRebate:  rebate,
			NetLossToUser:   loss.Sub(rebate),
			NetAmountToUser: zero,
		}
	default:
		return Result{
			Type:            TypeBreakeven,
			GrossPnL:        zero,
			GasFeeAmount:    zero,
			PlatformRebate:  zero,
			NetAmountToUser: zero,
			NetLossToUser:   zero,
		}
	}
}
