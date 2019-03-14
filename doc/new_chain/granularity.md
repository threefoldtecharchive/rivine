# Currency units

By default 1 coin = 1 000 000 000 of the smallest possible units. 
This can be overridden if desired by modifying the `ChainConstants`:

```
    cfg := types.StandardnetChainConstants()
	cfg.CurrencyUnits = types.CurrencyUnits{
		// 1 coin = 1 000  of the smalles possible units
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil)),
 ```
