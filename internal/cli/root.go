package cli

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "p2p-chat",
	Short: "p2p-chat is a decentralized private p2p chatting app",
	Long:  `A fully decentralized private p2p chatting application built with libp2p, LevelDB, and Svelte.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}


