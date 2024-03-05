package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GnarloqGames/genesis-avalon-daemon/config"
	"github.com/GnarloqGames/genesis-avalon-daemon/logging"
	"github.com/GnarloqGames/genesis-avalon-daemon/router"
	"github.com/GnarloqGames/genesis-avalon-daemon/worker"
	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
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

		bus, err := initMessageBus()
		if err != nil {
			return err
		}
		defer bus.Close()

		// Try connecting to Couchbase to catch issues at runtime
		if _, err := couchbase.Get(); err != nil {
			return err
		}

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

	rootCmd.PersistentFlags().String(config.FlagEnvironment, "development", "environment")
	rootCmd.PersistentFlags().String(config.FlagNatsAddress, "127.0.0.1:4222", "NATS address")
	rootCmd.PersistentFlags().String(config.FlagNatsEncoding, "json", "NATS encoding")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseURL, "127.0.0.1", "Couchbase host")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseBucket, "default", "Couchbase bucket")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseUsername, "", "Couchbase username")
	rootCmd.PersistentFlags().String(config.FlagCouchbasePassword, "", "Couchbase password")
	rootCmd.PersistentFlags().String(config.FlagLogLevel, "info", "log level (default is info)")
	rootCmd.PersistentFlags().String(config.FlagLogKind, "text", "log kind (text or json, default is text)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/gatewayd/config.yaml)")

	envPrefix := config.EnvPrefix
	bindFlags := map[string]string{
		config.FlagEnvironment:       config.EnvEnvironment,
		config.FlagLogLevel:          config.EnvLogLevel,
		config.FlagLogKind:           config.EnvLogKind,
		config.FlagNatsAddress:       config.EnvNatsAddress,
		config.FlagNatsEncoding:      config.EnvNatsEncoding,
		config.FlagCouchbaseURL:      config.EnvCouchbaseURL,
		config.FlagCouchbaseBucket:   config.EnvCouchbaseBucket,
		config.FlagCouchbaseUsername: config.EnvCouchbaseUsername,
		config.FlagCouchbasePassword: config.EnvCouchbasePassword,
	}

	for flag, env := range bindFlags {
		if err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
			slog.Warn("failed to bind flag", "error", err, "name", flag)
		}

		env = fmt.Sprintf("%s_%s", envPrefix, env)
		if err := viper.BindEnv(flag, env); err != nil {
			slog.Warn("failed to bind env", "error", err, "flag", flag, "env", env)
		}
	}

	viper.SetDefault(config.FlagLogLevel, "info")
	viper.SetDefault(config.FlagLogKind, "text")
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

	setConfigs()
}

func initMessageBus() (*transport.Connection, error) {
	natsAddress := viper.GetString(config.FlagNatsAddress)
	if natsAddress == "" {
		natsAddress = defaultNatsAddress
	}

	natsEncoder := viper.GetString(config.FlagNatsEncoding)
	if natsEncoder == "" {
		natsEncoder = defaultNatsEncoder
	}

	encoder := transport.ParseEncoder(natsEncoder)
	config := transport.DefaultConfig
	config.URL = natsAddress
	config.Encoder = encoder

	slog.Info("connecting to NATS service", "address", natsAddress, "encoder", natsEncoder)

	bus, err := transport.NewEncodedConn(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	slog.Info("established connection to NATS", "address", natsAddress, "encoding", natsEncoder)

	return bus, nil
}
