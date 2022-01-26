package cmd

import (
	"github.com/spf13/cobra"
)

const (
	backupsDir = "backups"
)

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupNewCmd())
	backupCmd.AddCommand(backupListCmd())
	backupCmd.AddCommand(backupRestoreCmd())
}

// backupCmd represents the balances command
var backupCmd = &cobra.Command{
	Use:          "backup",
	Short:        "Backup blockchain and wallet node",
	Long:         ``,
	SilenceUsage: true,
}

// backupNewCmd represents the balances list command
func backupNewCmd() *cobra.Command {
	var backupNewCmd = &cobra.Command{
		Use:   "new",
		Short: "Create versioned backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Warn("Not implemented: backup new called")
			return nil
		},
	}

	return backupNewCmd
}

// backupNewCmd represents the balances list command
func backupListCmd() *cobra.Command {
	var backupListCmd = &cobra.Command{
		Use:   "list",
		Short: "List backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Warn("Not implemented: backup list called")
			return nil
		},
	}

	addOutputFormatFlag(backupListCmd)

	return backupListCmd
}

// backupRestoreCmd represents the balances get command
func backupRestoreCmd() *cobra.Command {
	var backupRestoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Warn("Not implemented: backup list called")
			return nil
		},
	}

	return backupRestoreCmd
}
