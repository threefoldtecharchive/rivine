package client

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type (
	BaseClient struct {
		config     *Config
		httpClient HTTPClient
	}

	HTTPClient interface {
		Get(call string) error
		GetWithResponse(call string, obj interface{}) error

		Post(call, data string) error
		PostWithResponse(call, data string, obj interface{}) error
	}
)

func NewBaseClientFromCommandLineClient(cli *CommandLineClient) (*BaseClient, error) {
	return NewBaseClient(cli.HTTPClient, cli.Config)
}

func NewBaseClient(httpClient HTTPClient, config *Config) (*BaseClient, error) {
	if httpClient == nil {
		return nil, errors.New("no HTTPClient given while one is required for a base client")
	}
	if config == nil {
		var err error
		config, err = FetchConfigFromDaemon(httpClient)
		if err != nil {
			return nil, err
		}
	}
	return &BaseClient{
		config:     config,
		httpClient: httpClient,
	}, nil
}

func (bc *BaseClient) Config() *Config {
	return bc.config
}

func (bc *BaseClient) HTTP() HTTPClient {
	return bc.httpClient
}

// CreateCurrencyConvertor creates a currency convertor using the internally stored Config.
func (bc *BaseClient) CreateCurrencyConvertor() CurrencyConvertor {
	return NewCurrencyConvertor(
		bc.config.CurrencyUnits,
		bc.config.CurrencyCoinUnit,
	)
}

// FetchConfigFromDaemon fetches constants and creates a config, by fetching the constants from the daemon.
// Returns an error in case the fetching wasn't possible.
func FetchConfigFromDaemon(httpClient HTTPClient) (*Config, error) {
	if httpClient == nil {
		return nil, errors.New("cannot fetch config from daemon as no (HTTP) API client is configured")
	}

	var constants modules.DaemonConstants
	err := httpClient.GetWithResponse("/daemon/constants", &constants)
	if err == nil {
		// returned config from received constants from the daemon's server module
		cfg := ConfigFromDaemonConstants(constants)
		return &cfg, nil
	}
	err = httpClient.GetWithResponse("/explorer/constants", &constants)
	if err != nil {
		return nil, fmt.Errorf("failed to load constants from daemon's server and explorer modules: %v", err)
	}
	if constants.ChainInfo == (types.BlockchainInfo{}) {
		// only since 1.0.7 do we support the full set of public daemon constants for both
		// the explorer endpoint as well as the daemon endpoint,
		// so we need to validate this
		return nil, errors.New("failed to load constants from daemon's server and explorer modules: " +
			"explorer modules does not support the full exposure of public daemon constants")
	}
	// returned config from received constants from the daemon's explorer module
	cfg := ConfigFromDaemonConstants(constants)
	return &cfg, nil
}
