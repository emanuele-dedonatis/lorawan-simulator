package cli

import (
	"fmt"

	"github.com/brocaar/lorawan"
	"github.com/spf13/cobra"
)

var gwCmd = &cobra.Command{
	Use:   "gw <network-server>",
	Short: "Manage gateways in a network server",
	Args:  cobra.MinimumNArgs(1),
}

var gwAddCmd = &cobra.Command{
	Use:   "add <network-server> <eui> <discovery-uri>",
	Short: "Add a gateway to a network server",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		nsName := args[0]
		euiStr := args[1]
		discoveryURI := args[2]

		// Get the network server
		ns, exists := pool.Get(nsName)
		if !exists {
			return fmt.Errorf("network server %s not found", nsName)
		}

		// Parse the EUI
		var eui lorawan.EUI64
		if err := eui.UnmarshalText([]byte(euiStr)); err != nil {
			return fmt.Errorf("invalid EUI format: %v", err)
		}

		// Add the gateway
		if _, err := ns.AddGateway(eui, discoveryURI); err != nil {
			return err
		}

		fmt.Printf("gateway %s added to network server %s\n", eui, nsName)
		return nil
	},
}

var gwRemoveCmd = &cobra.Command{
	Use:   "remove <network-server> <eui>",
	Short: "Remove a gateway from a network server",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nsName := args[0]
		euiStr := args[1]

		// Get the network server
		ns, exists := pool.Get(nsName)
		if !exists {
			return fmt.Errorf("network server %s not found", nsName)
		}

		// Parse the EUI
		var eui lorawan.EUI64
		if err := eui.UnmarshalText([]byte(euiStr)); err != nil {
			return fmt.Errorf("invalid EUI format: %v", err)
		}

		// Remove the gateway
		if err := ns.RemoveGateway(eui); err != nil {
			return err
		}

		fmt.Printf("gateway %s removed from network server %s\n", eui, nsName)
		return nil
	},
}

var gwListCmd = &cobra.Command{
	Use:   "list <network-server>",
	Short: "List gateways in a network server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nsName := args[0]

		// Get the network server
		ns, exists := pool.Get(nsName)
		if !exists {
			return fmt.Errorf("network server %s not found", nsName)
		}

		// List gateways
		gateways := ns.ListGateways()
		if len(gateways) == 0 {
			fmt.Printf("no gateways in network server %s\n", nsName)
			return nil
		}

		fmt.Printf("Gateways in network server %s:\n", nsName)
		for _, gw := range gateways {
			info := gw.GetInfo()
			fmt.Printf("  %s | URI: %s | State: %s\n", info.EUI, info.DiscoveryURI, info.State)
		}
		return nil
	},
}

func InitGwCmd() *cobra.Command {
	gwCmd.AddCommand(gwAddCmd)
	gwCmd.AddCommand(gwRemoveCmd)
	gwCmd.AddCommand(gwListCmd)

	return gwCmd
}
