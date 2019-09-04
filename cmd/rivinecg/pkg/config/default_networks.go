package config

import (
	"reflect"
	"time"

	"github.com/threefoldtech/rivine/modules"
)

const (
	MaxPoolSize             uint = 2e7
	TransactionSetSizeLimit uint = 250e3
	TransactionSizeLimit    uint = 16e3
	ActualPoolSize               = uint64(MaxPoolSize - TransactionSetSizeLimit)
)

var (
	// networkRootPropsStandard are sane defaults for standard network configuration
	networkRootPropsStandard = map[string]interface{}{
		"BlockSizeLimit":         uint64(2e6),
		"ArbitraryDataSizeLimit": uint64(83),
		"BlockCreatorFee":        "1.0",
		"MinimumTransactionFee":  "0.1",
		"BlockFrequency":         uint64(120),
		"MaturityDelay":          uint64(144),
		"MedianTimestampWindow":  uint64(11),
		"TargetWindow":           uint64(1e3),
		"MaxAdjustmentUp":        Fraction{Denominator: 25, Numerator: 10},
		"MaxAdjustmentDown":      Fraction{Denominator: 25, Numerator: 10},
		"FutureThreshold":        uint64(time.Hour),
		"ExtremeFutureThreshold": uint64(time.Hour * 2),
		"StakeModifierDelay":     uint64(time.Second * 2000),
		"BlockStakeAging":        uint64(time.Hour * 24),
	}

	// networkRootPropsTestnet are sane defaults for testnet network configuration
	networkRootPropsTestnet = map[string]interface{}{
		"BlockSizeLimit":         uint64(2e6),
		"ArbitraryDataSizeLimit": uint64(83),
		"BlockCreatorFee":        "0.0",
		"MinimumTransactionFee":  "0.1",
		"BlockFrequency":         uint64(120),
		"MaturityDelay":          uint64(720),
		"MedianTimestampWindow":  uint64(11),
		"TargetWindow":           uint64(1e3),
		"MaxAdjustmentUp":        Fraction{Denominator: 25, Numerator: 10},
		"MaxAdjustmentDown":      Fraction{Denominator: 25, Numerator: 10},
		"FutureThreshold":        uint64(time.Second * 3),
		"ExtremeFutureThreshold": uint64(time.Second * 6),
		"StakeModifierDelay":     uint64(time.Second * 20),
		"BlockStakeAging":        uint64(time.Second * 1024),
	}

	// networkRootPropsDevnet are sane defaults for devnet network configuration
	networkRootPropsDevnet = map[string]interface{}{
		"BlockSizeLimit":         uint64(2e6),
		"ArbitraryDataSizeLimit": uint64(83),
		"BlockCreatorFee":        "0.0",
		"MinimumTransactionFee":  "0.1",
		"BlockFrequency":         uint64(12),
		"MaturityDelay":          uint64(10),
		"MedianTimestampWindow":  uint64(11),
		"TargetWindow":           uint64(20),
		"MaxAdjustmentUp":        Fraction{Denominator: 120, Numerator: 100},
		"MaxAdjustmentDown":      Fraction{Denominator: 100, Numerator: 120},
		"FutureThreshold":        uint64(time.Minute * 2),
		"ExtremeFutureThreshold": uint64(time.Minute * 3),
		"StakeModifierDelay":     uint64(time.Second * 2000),
		"BlockStakeAging":        uint64(time.Second * 1024),
		"BootstrapPeers": []modules.NetAddress{
			"localhost:23112",
		},
	}

	// networkTransactionPoolProps are sane defaults for standard network transactionPool configuration
	networkTransactionPoolProps = map[string]interface{}{
		"TransactionSizeLimit":    TransactionSizeLimit,
		"TransactionSetSizeLimit": TransactionSetSizeLimit,
		"PoolSizeLimit":           uint64(MaxPoolSize - TransactionSetSizeLimit),
	}
)

func assignDefaultNetworkProps(networkConfig *Network) {
	networkCfgValue := reflect.ValueOf(&networkConfig).Elem()
	var rootMap = map[string]interface{}{}
	switch networkConfig.NetworkType {
	case NetworkTypeStandard:
		rootMap = networkRootPropsStandard
	case NetworkTypeTestnet:
		rootMap = networkRootPropsTestnet
	case NetworkTypeDevnet:
		rootMap = networkRootPropsDevnet
	}

	for propName, propValue := range rootMap {
		pValue := networkCfgValue.Elem().FieldByName(propName)
		if isZero(pValue) {
			pValue.Set(reflect.ValueOf(propValue))
		}
	}
	networkTransactionPoolCfgValue := reflect.ValueOf(&networkConfig.TransactionPool)
	for propName, propValue := range networkTransactionPoolProps {
		pValue := networkTransactionPoolCfgValue.Elem().FieldByName(propName)
		if isZero(pValue) {
			pValue.Set(reflect.ValueOf(propValue))
		}
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				z = z && isZero(v.Field(i))
			}
		}
		return z
	case reflect.Ptr:
		return isZero(reflect.Indirect(v))
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	result := v.Interface() == z.Interface()

	return result
}
