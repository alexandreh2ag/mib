package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alexandreh2ag/mib/context"
)

const (
	Config     = "config"
	WorkingDir = "dir"
	LogLevel   = "level"
)

func GetRootCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "mib",
		Short:             "Manager multiple images with dependencies with other images",
		PersistentPreRunE: GetRootPreRunEFn(ctx),
	}

	cmd.PersistentFlags().StringP(Config, "c", "", "Define config path")
	cmd.PersistentFlags().StringP(WorkingDir, "w", "", "Define working dir")
	cmd.PersistentFlags().StringP(LogLevel, "l", "INFO", "Define log level")
	_ = viper.BindPFlag(Config, cmd.Flags().Lookup(Config))
	_ = viper.BindPFlag(WorkingDir, cmd.Flags().Lookup(WorkingDir))
	_ = viper.BindPFlag(LogLevel, cmd.Flags().Lookup(LogLevel))
	viper.SetDefault(LogLevel, "info")
	viper.RegisterAlias("log_level", LogLevel)

	return cmd
}

func GetRootPreRunEFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Lookup(WorkingDir).Value.String() != "" {
			workingDir, _ := cmd.Flags().GetString(WorkingDir)
			ctx.WorkingDir, _ = filepath.Abs(workingDir)
		}
		initConfig(ctx, cmd)
		logLevelFlagStr, _ := cmd.Flags().GetString(LogLevel)
		if logLevelFlagStr != "" {
			level := slog.LevelInfo
			err := level.UnmarshalText([]byte(logLevelFlagStr))
			if err != nil {
				return err
			}
			ctx.LogLevel.Set(level)
		}
		ctx.Logger.Info(fmt.Sprintf("Log level %s", ctx.LogLevel.String()))
		ctx.Logger.Info(fmt.Sprintf("Use working dir %s", ctx.WorkingDir))

		return nil
	}
}

func initConfig(ctx *context.Context, cmd *cobra.Command) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	viper.AddConfigPath(dir)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MIB")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Println(err)
	}

	configPath := viper.GetString(Config)

	if configPath != "" {
		viper.SetConfigFile(configPath)
		configDir := path.Dir(configPath)
		if configDir != "." && configDir != dir {
			viper.AddConfigPath(configDir)
		}

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			fmt.Println(err)
		}
	}

	err = viper.Unmarshal(ctx.Config)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

}
