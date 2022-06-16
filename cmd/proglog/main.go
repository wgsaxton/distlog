package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wgsaxton/distlog/internal/agent"
	"github.com/wgsaxton/distlog/internal/common"
	"github.com/wgsaxton/distlog/internal/config"
)

const ver string = "0.0.22"

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use: "proglog",
		// Use: "proglog [--config-file file]" ???
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}
	if err := setupFlags(cmd); err != nil {
		fmt.Println("setupFlags(cmd) returned error")
		log.Fatal(err)
	}
	if err := cmd.Execute(); err != nil {
		common.Gslog.Println("Error returned for main()-cmd.Execute() err:", err)
		log.Fatal(err)
	}
}

type cli struct {
	cfg cfg
}

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

func setupFlags(cmd *cobra.Command) error {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Version is", ver)
	common.Gslog.Println("Testing gs log. Version is:", ver)
	fmt.Println("Hostname found in setupFlags:", hostname)

	cmd.Flags().String("config-file", "", "Path to config file.")
	dataDir := path.Join(os.TempDir(), "proglog")
	cmd.Flags().String("data-dir",
		dataDir,
		"Directory to store log and Raft data.")
	cmd.Flags().String("node-name", hostname, "Unique server ID")
	cmd.Flags().String("bind-addr",
		"127.0.0.1:8401",
		"Address to bind Serf on")
	cmd.Flags().Int("rpc-port",
		8400,
		"Port for RPC clients (and Raft) connections.")
	cmd.Flags().StringSlice("start-join-addrs",
		nil,
		"Serf addresses to join.")
	cmd.Flags().Bool("bootstrap", false, "Bootstrap the cluster.")
	cmd.Flags().String("acl-model-file", "", "Path to ACL model.")
	cmd.Flags().String("acl-policy-file", "", "Path to ACL policy.")
	cmd.Flags().String("server-tls-cert-file", "", "Path to server tls cert.")
	cmd.Flags().String("server-tls-key-file", "", "Path to server tls key.")
	cmd.Flags().String("server-tls-ca-file",
		"",
		"Path to server certificate authority.")
	cmd.Flags().String("peer-tls-cert-file", "", "Path to peer tls cert.")
	cmd.Flags().String("peer-tls-key-file", "", "Path to peer tls key.")
	cmd.Flags().String("peer-tls-ca-file",
		"",
		"Path to peer cerificate authority.")
	cmd.Flags().Bool("version", false, "Version of Proglog")
	fmt.Println("GSnote - Done with setupFlags()")
	return viper.BindPFlags(cmd.Flags())
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	var err error

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	viper.SetConfigFile(configFile)
	fmt.Println("Config file set to:", viper.ConfigFileUsed())

	fmt.Println("Viper start reading config file.")
	if err = viper.ReadInConfig(); err != nil {
		// it's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("file found but an error")
			return err
		}
	}
	c.cfg.DataDir = viper.GetString("data-dir")
	common.Gslog.Println("c.cfg.DataDir value:", c.cfg.DataDir)
	c.cfg.NodeName = viper.GetString("node-name")
	c.cfg.BindAddr = viper.GetString("bind-addr")
	c.cfg.RPCPort = viper.GetInt("rpc-port")
	c.cfg.StartJoinAddrs = viper.GetStringSlice("start-join-addrs")
	c.cfg.Bootstrap = viper.GetBool("bootstrap")
	c.cfg.ACLModelFile = viper.GetString("acl-model-file")
	c.cfg.ACLPolicyFile = viper.GetString("acl-policy-file")
	c.cfg.ServerTLSConfig.CertFile = viper.GetString("server-tls-cert-file")
	c.cfg.ServerTLSConfig.KeyFile = viper.GetString("server-tls-key-file")
	c.cfg.ServerTLSConfig.CAFile = viper.GetString("server-tls-ca-file")
	c.cfg.PeerTLSConfig.CertFile = viper.GetString("peer-tls-cert-file")
	c.cfg.PeerTLSConfig.KeyFile = viper.GetString("peer-tls-key-file")
	c.cfg.PeerTLSConfig.CAFile = viper.GetString("peer-tls-ca-file")

	fmt.Printf("1st - The struct with values: %+v\n", c)

	if c.cfg.ServerTLSConfig.CertFile != "" &&
		c.cfg.ServerTLSConfig.KeyFile != "" {
		c.cfg.ServerTLSConfig.Server = true
		c.cfg.Config.ServerTLSConfig, err = config.SetupTLSConfig(
			c.cfg.ServerTLSConfig,
		)
		if err != nil {
			return err
		}
	}

	if c.cfg.PeerTLSConfig.CertFile != "" &&
		c.cfg.PeerTLSConfig.KeyFile != "" {
		c.cfg.Config.PeerTLSConfig, err = config.SetupTLSConfig(
			c.cfg.PeerTLSConfig,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("2nd - The struct with values: %+v\n", c)
	fmt.Println("Done with PreRunE: cli.setupConfig. No errors returned.")

	if viper.GetBool("version") {
		fmt.Println("Version is", ver)
		syscall.Exit(0)
	}
	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	var err error

	fmt.Println("proglog/main.go cli.run() starting")
	fmt.Printf("3rd - The struct with values: %+v\n", c)

	agent, err := agent.New(c.cfg.Config)
	if err != nil {
		common.Gslog.Println("Error given here. err:", err)
		return err
	}
	fmt.Println("Agent started")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	fmt.Println("Got shutdown signal")
	return agent.Shutdown()
}
