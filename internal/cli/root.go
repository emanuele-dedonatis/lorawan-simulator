package cli

import (
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/spf13/cobra"
)

var pool *networkserver.Pool

var rootCmd = &cobra.Command{
	Use:           "lorawan-simulator",
	Short:         "LoRaWAN simulator",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func InitRootCmd(p *networkserver.Pool) *cobra.Command {
	pool = p

	rootCmd.AddCommand(InitNsCmd())

	return rootCmd
}
