package electrum

import (
	"encoding/json"

	"github.com/rivine/rivine/types"
)

type (
	// AddressNotification is the info send to subscribers if the status
	AddressNotification struct {
		Address types.UnlockHash `json:"address"`
		Status  string           `json:"status"`
	}

	// rpcFunc is the signature for functions which can be registered for rpc invocations
	rpcFunc func(*Client, *json.RawMessage) (interface{}, error)
)

func (e *Electrum) registerServerMethods(cl *Client) error {
	var serverMethods = map[string]rpcFunc{
		"server.ping":    e.ServerPing,
		"server.version": e.ServerVersion,
	}
	for name, method := range serverMethods {
		if err := cl.registerService(name, method); err != nil {
			return err
		}
	}
	return nil
}

func (e *Electrum) registerBlockchainMethods(cl *Client) error {
	var blockchainMethods = map[string]rpcFunc{
		"blockchain.address.subscribe": e.BlockchainAddressSubscribe,
	}
	for name, method := range blockchainMethods {
		if err := cl.registerService(name, method); err != nil {
			return err
		}
	}
	return nil
}

// ServerPing is a utlitiy mehtod to refresh the connection timeout
// counter. It serves no other purpose. As a result, there are no input
// or output arguments
func (e *Electrum) ServerPing(cl *Client, args *json.RawMessage) (interface{}, error) {
	return nil, nil
}

// ServerVersion should be the first method called by a new connection. Before this call,
// only a very limit amount of calls are available.
func (e *Electrum) ServerVersion(cl *Client, args *json.RawMessage) (interface{}, error) {
	if (cl.protoVersion != ProtocolVersion{}) {
		// Protocol version already set, according to the spec new requests should be ignored
		// it is however not specified what "ignored" means. Completely ignore request, return
		// the information as set in the first request and don't do anything, or something else.
		//
		// For now return an error, can fix it later when version negotiation is properly implemented
		return nil, RPCError{
			Code:    ErrCodeVersionAlreadySet,
			Message: "Protocol version already set for this connection",
		}
	}
	// TODO: PROPER VERSION NEGOTIATION
	input := struct {
		ClientName      string           `json:"client_name"`
		ProtocolVersion ProtocolArgument `json:"protocol_version"`
	}{}

	resp := struct {
		ServerVersion string          `json:"server_software_version"`
		ProtoVersion  ProtocolVersion `json:"protocol_version"`
	}{}

	err := json.Unmarshal(*args, &input)
	if err != nil {
		e.log.Debug("Error getting params for server.version:", err)
		return nil, ErrInvalidParams
	}

	if input.ProtocolVersion.protocolMax != e.availableVersions[0] {
		// If no protocol version can be found, the connection
		// should be closed
		e.log.Debug("Error setting protocol, max version is " + input.ProtocolVersion.protocolMax.String() + " but we only accept 1.0 for now")
		return nil, errFatal
	}

	if err = e.registerBlockchainMethods(cl); err != nil {
		e.log.Println("[ERROR] Failed to register blockchain methods:", err)
		return nil, ErrInternal
	}

	e.log.Debug("Set proto version to ", input.ProtocolVersion.protocolMax.String(), " for client at ", cl.transport.RemoteAddr())
	resp.ProtoVersion = e.availableVersions[0]
	resp.ServerVersion = e.bcInfo.ChainVersion.String()

	cl.protoVersion = input.ProtocolVersion.protocolMax
	cl.clientName = input.ClientName

	return resp, nil
}

// BlockchainAddressSubscribe subscribes to a certain address
func (e *Electrum) BlockchainAddressSubscribe(cl *Client, args *json.RawMessage) (interface{}, error) {
	input := struct {
		Address types.UnlockHash `json:"address"`
	}{}

	resp := struct {
		Status string `json:"status"`
	}{}

	err := json.Unmarshal(*args, &input)
	if err != nil {
		e.log.Debug("Error getting params for blockhain.address.subscribe:", err)
		return nil, ErrInvalidParams
	}

	err = cl.registerAddressSubscription(input.Address)
	if err != nil && err != errAlreadySubscribed {
		e.log.Debug("Failed to register address subscription:", err)
	}

	resp.Status = e.AddressStatus(input.Address)

	return resp, err
}
