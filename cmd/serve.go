package cmd

import (
	"context"
	"nostar/internal/infrastrcture/db"
	"nostar/internal/relay/usecase"
	"nostar/internal/transport/websocket"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var servePort int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		zap.S().Infow("serve called", "port", servePort)

		// EventStore
		eventStore := db.NewEventStore()

		// RelayService
		relaySvc := usecase.NewRelayService(eventStore)

		// Server
		Srv := websocket.NewServer("0.0.0.0:9999", relaySvc)

		ctx := context.Background()
		_ = Srv.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "listen port")
}
