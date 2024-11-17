package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/richardamare/ai-digest/internal/config"
	"github.com/spf13/cobra"
)

var (
	configFile string
	configCmd  = &cobra.Command{
		Use:   "config",
		Short: "Manage AI Digest configuration",
		Long: `View and modify AI Digest configuration settings.
By default, configuration is stored in ai-digest.json in the current directory.`,
	}

	configShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE:  showConfig,
	}

	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		Long: `Initialize a new configuration file in the current directory.
If no config file exists, creates ai-digest.json with default settings.`,
		RunE: initConfig,
	}
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd, configInitCmd)

	// Make the config flag optional, default to CWD
	configCmd.PersistentFlags().StringVar(&configFile, "config", "",
		"config file path (defaults to ./ai-digest.json)")
}

func showConfig(cmd *cobra.Command, args []string) error {
	manager := config.NewManager(configFile)

	if !manager.Exists() {
		fmt.Println("No configuration file found. Using default settings:")
	}

	output, err := manager.Show()
	if err != nil {
		return fmt.Errorf("failed to show config: %w", err)
	}

	fmt.Println(output)
	return nil
}

func initConfig(cmd *cobra.Command, args []string) error {
	manager := config.NewManager(configFile)

	if manager.Exists() {
		return fmt.Errorf("config file already exists at %s", manager.GetConfigPath())
	}

	if err := manager.Init(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	absPath, err := filepath.Abs(manager.GetConfigPath())
	if err != nil {
		absPath = manager.GetConfigPath()
	}

	fmt.Printf("Created config file: %s\n", absPath)
	fmt.Println("You can now modify this file or use 'ai-digest config show' to view it")
	return nil
}
