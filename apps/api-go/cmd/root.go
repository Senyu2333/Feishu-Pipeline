package cmd

import "github.com/spf13/cobra"

var version = "dev"

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "requirement-delivery-api",
		Short: "需求交付流程引擎服务",
	}

	root.AddCommand(newServeCommand())
	root.Version = version
	return root
}
