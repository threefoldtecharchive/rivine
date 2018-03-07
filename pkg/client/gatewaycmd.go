package client

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/rivine/rivine/api"
)

var (
	gatewayCmd = &cobra.Command{
		Use:   "gateway",
		Short: "Perform gateway actions",
		Long:  "View and manage the gateway's connected peers.",
		Run:   wrap(Gatewaycmd),
	}

	gatewayConnectCmd = &cobra.Command{
		Use:   "connect [address]",
		Short: "Connect to a peer",
		Long:  "Connect to a peer and add it to the node list.",
		Run:   wrap(Gatewayconnectcmd),
	}

	gatewayDisconnectCmd = &cobra.Command{
		Use:   "disconnect [address]",
		Short: "Disconnect from a peer",
		Long:  "Disconnect from a peer. Does not remove the peer from the node list.",
		Run:   wrap(Gatewaydisconnectcmd),
	}

	gatewayAddressCmd = &cobra.Command{
		Use:   "address",
		Short: "Print the gateway address",
		Long:  "Print the network address of the gateway.",
		Run:   wrap(Gatewayaddresscmd),
	}

	gatewayListCmd = &cobra.Command{
		Use:   "list",
		Short: "View a list of peers",
		Long:  "View the current peer list.",
		Run:   wrap(Gatewaylistcmd),
	}
)

// Gatewayconnectcmd is the handler for the command `siac gateway add [address]`.
// Adds a new peer to the peer list.
func Gatewayconnectcmd(addr string) {
	err := Post("/gateway/connect/"+addr, "")
	if err != nil {
		Die("Could not add peer:", err)
	}
	fmt.Println("Added", addr, "to peer list.")
}

// Gatewaydisconnectcmd is the handler for the command `siac gateway remove [address]`.
// Removes a peer from the peer list.
func Gatewaydisconnectcmd(addr string) {
	err := Post("/gateway/disconnect/"+addr, "")
	if err != nil {
		Die("Could not remove peer:", err)
	}
	fmt.Println("Removed", addr, "from peer list.")
}

// Gatewayaddresscmd is the handler for the command `siac gateway address`.
// Prints the gateway's network address.
func Gatewayaddresscmd() {
	var info api.GatewayGET
	err := GetAPI("/gateway", &info)
	if err != nil {
		Die("Could not get gateway address:", err)
	}
	fmt.Println("Address:", info.NetAddress)
}

// Gatewaycmd is the handler for the command `siac gateway`.
// Prints the gateway's network address and number of peers.
func Gatewaycmd() {
	var info api.GatewayGET
	err := GetAPI("/gateway", &info)
	if err != nil {
		Die("Could not get gateway address:", err)
	}
	fmt.Println("Address:", info.NetAddress)
	fmt.Println("Active peers:", len(info.Peers))
}

// Gatewaylistcmd is the handler for the command `siac gateway list`.
// Prints a list of all peers.
func Gatewaylistcmd() {
	var info api.GatewayGET
	err := GetAPI("/gateway", &info)
	if err != nil {
		Die("Could not get peer list:", err)
	}
	if len(info.Peers) == 0 {
		fmt.Println("No peers to show.")
		return
	}
	fmt.Println(len(info.Peers), "active peers:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tOutbound\tAddress")
	for _, peer := range info.Peers {
		fmt.Fprintf(w, "%v\t%v\t%v\n", peer.Version, YesNo(!peer.Inbound), peer.NetAddress)
	}
	w.Flush()
}
