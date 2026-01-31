package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nsCmd = &cobra.Command{
	Use:   "ns",
	Short: "Manage network servers",
}

var nsAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a network server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if _, err := pool.Add(name); err != nil {
			return err
		}

		fmt.Printf("network server %s added\n", name)
		return nil
	},
}

var nsRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a network server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := pool.Remove(name); err != nil {
			return err
		}

		fmt.Printf("network server %s removed\n", name)
		return nil
	},
}

var nsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List network servers",
	Run: func(cmd *cobra.Command, args []string) {
		for _, ns := range pool.List() {
			fmt.Printf("%s | Devices: %d | Gateways: %d\n",
				ns.Name, ns.DeviceCount, ns.GatewayCount)
		}
	},
}

func InitNsCmd() *cobra.Command {
	nsCmd.AddCommand(nsAddCmd)
	nsCmd.AddCommand(nsRemoveCmd)
	nsCmd.AddCommand(nsListCmd)

	return nsCmd
}
