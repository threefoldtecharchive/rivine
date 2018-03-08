package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/build"
	"github.com/spf13/cobra"
)

// flags
var (
	addr string // override default API address
)

var (
	// ClientName sets the client name for some of the command help messages
	ClientName = "rivine"
)

// exit codes
// inspired by sysexits.h
const (
	exitCodeGeneral = 1  // Not in sysexits.h, but is standard practice.
	exitCodeUsage   = 64 // EX_USAGE in sysexits.h
)

// Non2xx returns true for non-success HTTP status codes.
func Non2xx(code int) bool {
	return code < 200 || code > 299
}

// DecodeError returns the api.Error from a API response. This method should
// only be called if the response's status code is non-2xx. The error returned
// may not be of type api.Error in the event of an error unmarshalling the
// JSON.
func DecodeError(resp *http.Response) error {
	var apiErr api.Error
	err := json.NewDecoder(resp.Body).Decode(&apiErr)
	if err != nil {
		return err
	}
	return apiErr
}

// ApiGet wraps a GET request with a status code check, such that if the GET does
// not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func ApiGet(call string) (*http.Response, error) {
	if host, port, _ := net.SplitHostPort(addr); host == "" {
		addr = net.JoinHostPort("localhost", port)
	}
	resp, err := api.HttpGET("http://" + addr + call)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// Prompt for password and retry request with authentication.
		password, err := speakeasy.Ask("API password: ")
		if err != nil {
			return nil, err
		}
		resp, err = api.HttpGETAuthenticated("http://"+addr+call, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errors.New("API call not recognized: " + call)
	}
	if Non2xx(resp.StatusCode) {
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// GetAPI makes a GET API call and decodes the response. An error is returned
// if the response status is not 2xx.
func GetAPI(call string, obj interface{}) error {
	resp, err := ApiGet(call)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return errors.New("expecting a response, but API returned status code 204 No Content")
	}

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// Get makes an API call and discards the response. An error is returned if the
// response status is not 2xx.
func Get(call string) error {
	resp, err := ApiGet(call)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ApiPost wraps a POST request with a status code check, such that if the POST
// does not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func ApiPost(call, vals string) (*http.Response, error) {
	if host, port, _ := net.SplitHostPort(addr); host == "" {
		addr = net.JoinHostPort("localhost", port)
	}

	resp, err := api.HttpPOST("http://"+addr+call, vals)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// Prompt for password and retry request with authentication.
		password, err := speakeasy.Ask("API password: ")
		if err != nil {
			return nil, err
		}
		resp, err = api.HttpPOSTAuthenticated("http://"+addr+call, vals, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errors.New("API call not recognized: " + call)
	}
	if Non2xx(resp.StatusCode) {
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// PostResp makes a POST API call and decodes the response. An error is
// returned if the response status is not 2xx.
func PostResp(call, vals string, obj interface{}) error {
	resp, err := ApiPost(call, vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return errors.New("expecting a response, but API returned status code 204 No Content")
	}

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// Post makes an API call and discards the response. An error is returned if
// the response status is not 2xx.
func Post(call, vals string) error {
	resp, err := ApiPost(call, vals)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// wrap wraps a generic command with a check that the command has been
// passed the correct number of arguments. The command must take only strings
// as arguments.
func wrap(fn interface{}) func(*cobra.Command, []string) {
	fnVal, fnType := reflect.ValueOf(fn), reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("wrapped function has wrong type signature")
	}
	for i := 0; i < fnType.NumIn(); i++ {
		if fnType.In(i).Kind() != reflect.String {
			panic("wrapped function has wrong type signature")
		}
	}

	return func(cmd *cobra.Command, args []string) {
		if len(args) != fnType.NumIn() {
			cmd.UsageFunc()(cmd)
			os.Exit(exitCodeUsage)
		}
		argVals := make([]reflect.Value, fnType.NumIn())
		for i := range args {
			argVals[i] = reflect.ValueOf(args[i])
		}
		fnVal.Call(argVals)
	}
}

// Die prints its arguments to stderr, then exits the program with the default
// error code.
func Die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(exitCodeGeneral)
}

// Version prints the client version and exits
func Version() {
	println(fmt.Sprintf("%s Client v", strings.Title(ClientName)) + build.Version.String())
}

// DefaultClient parses the arguments using cobra with the default rivine setup
func DefaultClient() {
	root := &cobra.Command{
		Use:   os.Args[0],
		Short: fmt.Sprintf("%s Client v", strings.Title(ClientName)) + build.Version.String(),
		Long:  fmt.Sprintf("%s Client v", strings.Title(ClientName)) + build.Version.String(),
		Run:   wrap(Consensuscmd),
	}

	// create command tree
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information.",
		Run:   wrap(Version),
	})

	root.AddCommand(stopCmd)

	root.AddCommand(updateCmd)
	updateCmd.AddCommand(updateCheckCmd)

	root.AddCommand(walletCmd)
	walletCmd.AddCommand(
		walletAddressCmd,
		walletAddressesCmd,
		walletInitCmd,
		walletLoadCmd,
		walletLockCmd,
		walletSeedsCmd,
		walletSendCmd,
		walletBalanceCmd,
		walletTransactionsCmd,
		walletUnlockCmd,
		walletBlockStakeStatCmd,
		walletRegisterDataCmd)

	walletSendCmd.AddCommand(
		walletSendSiacoinsCmd,
		walletSendSiafundsCmd)

	walletLoadCmd.AddCommand(walletLoadSeedCmd)

	root.AddCommand(gatewayCmd)
	gatewayCmd.AddCommand(
		gatewayConnectCmd,
		gatewayDisconnectCmd,
		gatewayAddressCmd,
		gatewayListCmd)

	root.AddCommand(consensusCmd)

	// parse flags
	root.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:23110", fmt.Sprintf("which host/port to communicate with (i.e. the host/port %sd is listening on)", ClientName))

	// run
	if err := root.Execute(); err != nil {
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(exitCodeUsage)
	}
}
