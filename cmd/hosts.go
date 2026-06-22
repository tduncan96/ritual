package cmd

import (
	"fmt"
	"log/slog"
	"strconv"

	"ritual/internal/db"

	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "manage hosts for ssh executions",
}

var addHost = &cobra.Command{
	Use:   "add <hostname> <ip address> <user> <port; if none default to 22> <key-path; if none default to ~/.ssh/id_ed25519>",
	Short: "add a host for job execution",
	Args:  cobra.RangeArgs(3, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		newHost := db.Host{
			Name:    args[0],
			Address: args[1],
			User:    args[2],
		}
		if len(args) > 3 {
			port, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}
			newHost.Port = int64(port)
		}
		if len(args) > 4 {
			newHost.KeyPath = args[4]
		}

		id, err := newHost.AddHost()
		if err != nil {
			return err
		}
		slog.Info("host created", "id", id, "name", newHost.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "host %v successfully added: ID: %d\n", newHost.Name, id)
		return nil
	},
}

func init() {
	hostCmd.AddCommand(addHost)

	rootCmd.AddCommand(hostCmd)
}
