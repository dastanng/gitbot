package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/dastanng/gitbot/pkg/bot"
)

var (
	opts    bot.InitOptions
	rootCmd = &cobra.Command{
		Use:          "bot",
		Short:        "github bot",
		Long:         "github bot watches github events and reacts respectively.",
		SilenceUsage: true,
	}
	webhookCmd = &cobra.Command{
		Use:   "webhook",
		Short: "Start webhook service",
		RunE: func(*cobra.Command, []string) error {
			logs.InitLogs()
			defer logs.FlushLogs()
			flag.CommandLine.Parse([]string{})

			bot := new(bot.Bot)
			bot.Initialize(opts)

			stopCh := setupSignalHandler()
			bot.Run(stopCh)
			return nil
		},
	}
)

func init() {
	webhookCmd.PersistentFlags().StringVar(&opts.Token, "token", "",
		"A token that can be used to access the GitHub API")
	cobra.MarkFlagRequired(
		webhookCmd.PersistentFlags(),
		"token",
	)
	webhookCmd.PersistentFlags().StringVar(&opts.Secret, "secret", "",
		"A secret that is used to validate the GitHub Webhook requests")
	cobra.MarkFlagRequired(
		webhookCmd.PersistentFlags(),
		"secret",
	)
	rootCmd.AddCommand(webhookCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

// setupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		close(stop)
		<-sigs
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}
