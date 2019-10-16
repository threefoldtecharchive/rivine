package client

import (
	"errors"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type (
	BaseClient interface {
		Config() (*Config, error)
		HTTP() HTTPClient
		NewCurrencyConvertor() (CurrencyConvertor, error)
	}

	LazyBaseClient struct {
		once func() (BaseClient, error)
		so   sync.Once
		bc   BaseClient
		err  error
	}

	StdBaseClient struct {
		config     *Config
		httpClient HTTPClient
	}

	HTTPClient interface {
		Get(call string) error
		GetWithResponse(call string, obj interface{}) error

		Post(call, data string) error
		PostWithResponse(call, data string, obj interface{}) error
	}

	errHTTPClient struct {
		err error
	}
)

func NewLazyBaseClientFromCommandLineClient(cli *CommandLineClient) (*LazyBaseClient, error) {
	return &LazyBaseClient{
		once: func() (BaseClient, error) {
			return NewBaseClientFromCommandLineClient(cli)
		},
	}, nil
}

func NewLazyBaseClient(once func() (BaseClient, error)) (*LazyBaseClient, error) {
	return &LazyBaseClient{
		once: once,
	}, nil
}

func (lbc *LazyBaseClient) Config() (*Config, error) {
	lbc.so.Do(lbc.doOnceFn)
	if lbc.err != nil {
		return nil, lbc.err
	}
	return lbc.bc.Config()
}

func (lbc *LazyBaseClient) HTTP() HTTPClient {
	lbc.so.Do(lbc.doOnceFn)
	if lbc.err != nil {
		return newErrHTTPClient(lbc.err)
	}
	return lbc.bc.HTTP()
}

func (lbc *LazyBaseClient) NewCurrencyConvertor() (CurrencyConvertor, error) {
	lbc.so.Do(lbc.doOnceFn)
	if lbc.err != nil {
		return CurrencyConvertor{}, lbc.err
	}
	return lbc.bc.NewCurrencyConvertor()
}

func (lbc *LazyBaseClient) doOnceFn() {
	lbc.bc, lbc.err = lbc.once()
}

func NewBaseClientFromCommandLineClient(cli *CommandLineClient) (BaseClient, error) {
	return NewBaseClient(cli.HTTPClient, cli.Config)
}

func NewBaseClient(httpClient HTTPClient, config *Config) (BaseClient, error) {
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
	return &StdBaseClient{
		config:     config,
		httpClient: httpClient,
	}, nil
}

func (bc *StdBaseClient) Config() (*Config, error) {
	return bc.config, nil
}

func (bc *StdBaseClient) HTTP() HTTPClient {
	return bc.httpClient
}

// NewCurrencyConvertor creates a currency convertor using the internally stored Config.
func (bc *StdBaseClient) NewCurrencyConvertor() (CurrencyConvertor, error) {
	return NewCurrencyConvertor(
		bc.config.CurrencyUnits,
		bc.config.CurrencyCoinUnit,
	), nil
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

func newErrHTTPClient(err error) *errHTTPClient {
	return &errHTTPClient{
		err: fmt.Errorf("failed to create HTTP Client: %v", err),
	}
}

func (ec *errHTTPClient) Get(call string) error {
	return fmt.Errorf("GET %s failed: %v", call, ec.err)
}

func (ec *errHTTPClient) GetWithResponse(call string, obj interface{}) error {
	return fmt.Errorf("GET (/wr) %s failed: %v", call, ec.err)
}

func (ec *errHTTPClient) Post(call, data string) error {
	return fmt.Errorf("POST %s failed: %v", call, ec.err)
}

func (ec *errHTTPClient) PostWithResponse(call, data string, obj interface{}) error {
	return fmt.Errorf("POST (/wr) %s failed: %v", call, ec.err)
}
