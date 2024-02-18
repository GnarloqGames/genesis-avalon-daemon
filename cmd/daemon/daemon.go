package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GnarloqGames/genesis-avalon-daemon/logging"
	"github.com/GnarloqGames/genesis-avalon-daemon/router"
	"github.com/GnarloqGames/genesis-avalon-daemon/worker"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

const (
	defaultNatsAddress string = "127.0.0.1:4222"
	defaultNatsEncoder string = "protobuf"
)

var rootCmd = &cobra.Command{
	Use:   "avalond",
	Short: "The daemon responsible for executing game logic",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the game daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

		bus, err := initMessageBus(cmd)
		if err != nil {
			return err
		}
		defer bus.Close()

		pool := worker.NewSystem()

		router.New(bus, pool)

		<-stopChan
		slog.Info("shutting down daemon")

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	startCmd.Flags().String("nats-address", "127.0.0.1:4222", "NATS address")
	startCmd.Flags().String("nats-encoding", "json", "NATS encoding")

	rootCmd.PersistentFlags().String("log-level", "info", "log level (default is info)")
	rootCmd.PersistentFlags().String("log-kind", "text", "log kind (text or json, default is text)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/gatewayd/config.yaml)")
	viper.SetDefault("log-level", "info")
	viper.SetDefault("log-kind", "text")
	viper.SetDefault("author", "Alfred Dobradi <alfreddobradi@gmail.com>")
	viper.SetDefault("license", "MIT")

	rootCmd.AddCommand(startCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.AddConfigPath("/etc/avalond")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("loaded config file", "file", viper.ConfigFileUsed())
	}

	if err := logging.Init(); err != nil {
		slog.Error("failed to create logger", "error", err.Error())
	}
}

func initMessageBus(cmd *cobra.Command) (*transport.Connection, error) {
	natsAddress, err := cmd.Flags().GetString("nats-address")
	if err != nil {
		natsAddress = defaultNatsAddress
	}

	natsEncoder, err := cmd.Flags().GetString("nats-encoder")
	if err != nil {
		natsEncoder = defaultNatsEncoder
	}

	encoder := transport.ParseEncoder(natsEncoder)
	config := transport.DefaultConfig
	config.URL = natsAddress
	config.Encoder = encoder

	bus, err := transport.NewEncodedConn(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	slog.Info("established connection to NATS", "address", natsAddress, "encoding", natsEncoder)

	return bus, nil
}
