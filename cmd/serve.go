package cmd

import (
	"context"
	"fmt"
	"nostar/internal/infrastructure/db"
	"nostar/internal/relay/usecase"
	"nostar/internal/transport/websocket"
	"os"

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

		ctx := context.Background()

		// DB connection (check at startup)
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			zap.S().Error("DATABASE_URL is not set")
			os.Exit(1)
			return
		}

		gormDB, err := db.NewGormDB(ctx, db.Config{DSN: dsn})
		if err != nil {
			zap.S().Errorw("failed to connect database", "error", err)
			os.Exit(1)
			return
		}

		// EventStore
		eventStore := db.NewEventStore(gormDB)

		// RelayService
		relaySvc := usecase.NewRelayService(eventStore)

		// Server
		addr := fmt.Sprintf("0.0.0.0:%d", servePort)
		Srv := websocket.NewServer(addr, relaySvc)

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
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 9999, "listen port")
}
