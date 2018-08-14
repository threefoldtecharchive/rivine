package client

import (
	"fmt"
	"os"

	"github.com/rivine/rivine/types"
)

func parseCoinArg(cc CurrencyConvertor, str string) types.Currency {
	amount, err := cc.ParseCoinString(str)
	if err != nil {
		fmt.Fprintln(os.Stderr, cc.CoinArgDescription("amount"))
		DieWithExitCode(ExitCodeUsage, "failed to parse coin-typed argument: ", err)
		return types.Currency{}
	}
	return amount
}
