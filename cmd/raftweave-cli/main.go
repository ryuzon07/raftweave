package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "raftweave",
		Short: "RaftWeave CLI — sovereign multi-cloud orchestration",
		Long:  "RaftWeave CLI provides commands to manage workloads, monitor cluster state, and trigger failovers across AWS, Azure, and GCP.",
	}

	rootCmd.AddCommand(
		newApplyCmd(),
		newStatusCmd(),
		newLogsCmd(),
		newFailoverCmd(),
		newCloudCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newApplyCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Submit a workload descriptor",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "applying workload from %s\n", file)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to workload descriptor YAML")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [workload-name]",
		Short: "Get cluster status for a workload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "status for workload: %s\n", args[0])
			return nil
		},
	}
}

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs [workload-name]",
		Short: "Stream build/runtime logs for a workload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "streaming logs for: %s\n", args[0])
			return nil
		},
	}
}

func newFailoverCmd() *cobra.Command {
	var toRegion string
	cmd := &cobra.Command{
		Use:   "failover [workload-name]",
		Short: "Trigger a manual failover for a workload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "triggering failover for %s to %s\n", args[0], toRegion)
			return nil
		},
	}
	cmd.Flags().StringVar(&toRegion, "to", "", "Target region for failover")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func newCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud provider credentials",
	}

	addCmd := &cobra.Command{
		Use:   "add [provider]",
		Short: "Add cloud credentials for a provider (aws, azure, gcp)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "adding cloud credentials for: %s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(addCmd)
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "raftweave version %s\n", version)
		},
	}
}
