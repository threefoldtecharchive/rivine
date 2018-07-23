package electrum

import (
	"errors"
	"sync"
	"time"

	"github.com/rivine/rivine/types"
)

type (
	// Client is a connection on which the electrum RPC protocol is served.
	Client struct {
		transport RPCTransport

		mu         sync.RWMutex
		serviceMap map[string]rpcFunc
		timer      *time.Timer

		addressSubscriptions map[types.UnlockHash]bool

		clientName   string
		protoVersion ProtocolVersion
	}
)

func (cl *Client) registerService(name string, f rpcFunc) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if _, exists := cl.serviceMap[name]; exists {
		return errors.New("A function with name " + name + " already exists for this connection")
	}
	cl.serviceMap[name] = f
	return nil
}

func (cl *Client) sendUpdate(update *Update) {
	for subscribedAddress := range cl.addressSubscriptions {
		if status, exists := update.addressStates[subscribedAddress]; exists {
			cl.notify("blockchain.address.subscribe",
				AddressNotification{Address: subscribedAddress, Status: status})
		}
	}
}

func (cl *Client) notify(method string, params interface{}) {
	n := &Notification{
		JSONRPC: jsonRPCVersion,
		Method:  method,
		Params:  params,
	}

	cl.transport.Send(n)
}

func (cl *Client) registerAddressSubscription(address types.UnlockHash) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if _, exists := cl.addressSubscriptions[address]; exists {
		return errors.New("Already subscribed to this address")
	}

	cl.addressSubscriptions[address] = true
	return nil
}

func (cl *Client) call(r *Request) (interface{}, error) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	f, exists := cl.serviceMap[r.Method]
	if !exists {
		return nil, ErrMethodNotFound
	}

	return f(cl, r.Params)
	// result, err := f(cl, r.Params)
	// response := newResponse(r.ID, result, err)

	// // If response is nil there is nothing to send, so return here
	// if response == nil {
	// 	return nil
	// }
	// cl.transport.Send(response)
	// return nil

}

// resetTimer makes sure the timer is reset properly
func (cl *Client) resetTimer() {
	// Reset must be done on a stopped or expired timer to be thread safe
	if !cl.timer.Stop() {
		// Try to drain the channel in case it fired just before the reset
		select {
		case <-cl.timer.C:
		default:
		}
	}
	cl.timer.Reset(connectionTimeout)
}
