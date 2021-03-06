package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd                        = &cobra.Command{Use: "ocidist"}
	username, password, proxyUrl   string
	anonymous, httpClient, verbose bool
)

func init() {
	rootCmd.AddCommand(pullCmd)
	pullInit()
	rootCmd.AddCommand(pushCmd)
	pushInit()
	rootCmd.AddCommand(copyCmd)
	copyInit()
	rootCmd.AddCommand(convertCmd)
	convertInit()
	rootCmd.AddCommand(mergeImageCmd)
	mergeImageInit()

	rootCmd.PersistentFlags().StringVar(&username, "username", "", "username to authenticate against registry")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "password to authenticate against registry")
	rootCmd.PersistentFlags().BoolVar(&anonymous, "anonymous", false, "use anonymous auth, defaults to your local credentials")
	rootCmd.PersistentFlags().BoolVar(&httpClient, "http", false, "use our own http client, rather than the default, required for proxy")
	rootCmd.PersistentFlags().StringVar(&proxyUrl, "proxy", "", "proxy URL to use")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "print lots of output to stderr")
}

// Execute primary function for cobra
func Execute() {
	rootCmd.Execute()
}
