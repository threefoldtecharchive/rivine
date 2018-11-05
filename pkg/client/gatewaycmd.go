package client

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/spf13/cobra"
)

func createGatewayCmd(cli *CommandLineClient) *cobra.Command {
	gatewayCmd := &gatewayCmd{cli: cli}

	// create root gateway command and all subs
	var (
		rootCmd = &cobra.Command{
			Use:   "gateway",
			Short: "Perform gateway actions",
			Long:  "View and manage the gateway's connected peers.",
			Run:   Wrap(gatewayCmd.rootCmd),
		}
		connectCmd = &cobra.Command{
			Use:   "connect [address]",
			Short: "Connect to a peer",
			Long:  "Connect to a peer and add it to the node list.",
			Run:   Wrap(gatewayCmd.connectCmd),
		}
		disconnectCmd = &cobra.Command{
			Use:   "disconnect [address]",
			Short: "Disconnect from a peer",
			Long:  "Disconnect from a peer. Does not remove the peer from the node list.",
			Run:   Wrap(gatewayCmd.disconnectCmd),
		}
		addressCmd = &cobra.Command{
			Use:   "address",
			Short: "Print the gateway address",
			Long:  "Print the network address of the gateway.",
			Run:   Wrap(gatewayCmd.addressCmd),
		}
		listPeersCmd = &cobra.Command{
			Use:   "list",
			Short: "View a list of peers",
			Long:  "View the current peer list.",
			Run:   Wrap(gatewayCmd.listPeersCmd),
		}
	)
	rootCmd.AddCommand(
		connectCmd,
		disconnectCmd,
		addressCmd,
		listPeersCmd,
	)

	// return root command
	return rootCmd
}

type gatewayCmd struct {
	cli *CommandLineClient
}

// connectCmd is the handler for the command `gateway add [address]`.
// Adds a new peer to the peer list.
func (gatewayCmd *gatewayCmd) connectCmd(addr string) {
	err := gatewayCmd.cli.Post("/gateway/connect/"+addr, "")
	if err != nil {
		cli.Die("Could not add peer:", err)
	}
	fmt.Println("Added", addr, "to peer list.")
}

// disconnectCmd is the handler for the command `gateway remove [address]`.
// Removes a peer from the peer list.
func (gatewayCmd *gatewayCmd) disconnectCmd(addr string) {
	err := gatewayCmd.cli.Post("/gateway/disconnect/"+addr, "")
	if err != nil {
		cli.Die("Could not remove peer:", err)
	}
	fmt.Println("Removed", addr, "from peer list.")
}

// addressCmd is the handler for the command `gateway address`.
// Prints the gateway's network address.
func (gatewayCmd *gatewayCmd) addressCmd() {
	var info api.GatewayGET
	err := gatewayCmd.cli.GetAPI("/gateway", &info)
	if err != nil {
		cli.Die("Could not get gateway address:", err)
	}
	fmt.Println("Address:", info.NetAddress)
}

// rootCmd is the handler for the command `gateway`.
// Prints the gateway's network address and number of peers.
func (gatewayCmd *gatewayCmd) rootCmd() {
	var info api.GatewayGET
	err := gatewayCmd.cli.GetAPI("/gateway", &info)
	if err != nil {
		cli.Die("Could not get gateway address:", err)
	}
	fmt.Println("Address:", info.NetAddress)
	fmt.Println("Active peers:", len(info.Peers))
}

// listPeersCmd is the handler for the command `gateway list`.
// Prints a list of all peers.
func (gatewayCmd *gatewayCmd) listPeersCmd() {
	var info api.GatewayGET
	err := gatewayCmd.cli.GetAPI("/gateway", &info)
	if err != nil {
		cli.Die("Could not get peer list:", err)
	}
	if len(info.Peers) == 0 {
		fmt.Println("No peers to show.")
		return
	}
	fmt.Println(len(info.Peers), "active peers:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tOutbound\tAddress")
	for _, peer := range info.Peers {
		fmt.Fprintf(w, "%s\t%v\t%v\n", peer.Version, YesNo(!peer.Inbound), peer.NetAddress)
	}
	w.Flush()
}
