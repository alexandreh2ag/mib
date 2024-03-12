package cli

import (
	"fmt"
	"github.com/alexandreh2ag/mib/container"
	"github.com/alexandreh2ag/mib/template"
	"log/slog"
	"path"
	"path/filepath"

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
	cmd.PersistentFlags().StringP(WorkingDir, "w", ctx.WorkingDir, "Define working dir")
	cmd.PersistentFlags().StringP(LogLevel, "l", "INFO", "Define log level")
	_ = viper.BindPFlag(Config, cmd.Flags().Lookup(Config))
	_ = viper.BindPFlag(WorkingDir, cmd.Flags().Lookup(WorkingDir))
	_ = viper.BindPFlag(LogLevel, cmd.Flags().Lookup(LogLevel))
	viper.SetDefault(LogLevel, "info")
	viper.RegisterAlias("log_level", LogLevel)
	cmd.AddCommand(
		GetBuildCmd(ctx),
		GetGenerateCmd(ctx),
		GetListCmd(ctx),
		GetCommitCmd(ctx),
		GetVersionCmd(),
	)
	return cmd
}

func GetRootPreRunEFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		workingDirFlag, err := cmd.Flags().GetString(WorkingDir)
		if err == nil && workingDirFlag != "" {
			workingDir, _ := cmd.Flags().GetString(WorkingDir)
			ctx.WorkingDir, _ = filepath.Abs(workingDir)
		}
		initConfig(ctx, cmd)
		err = template.OverrideTemplatesFromConfig(ctx)
		if err != nil {
			return err
		}
		logLevelFlagStr, _ := cmd.Flags().GetString(LogLevel)
		if logLevelFlagStr != "" && cmd.Flags().Changed(LogLevel) {
			level := slog.LevelInfo
			err := level.UnmarshalText([]byte(logLevelFlagStr))
			if err != nil {
				return err
			}
			ctx.LogLevel.Set(level)
		}
		ctx.Logger.Info(fmt.Sprintf("Log level %s", ctx.LogLevel.String()))
		ctx.Logger.Info(fmt.Sprintf("Use working dir %s", ctx.WorkingDir))

		for key, fn := range container.BuilderFnFactory {
			builder, errCreateBuilder := fn(ctx)
			if errCreateBuilder != nil {
				return fmt.Errorf("fail to create builder %s with error: %v", key, errCreateBuilder)
			}
			ctx.Builders[key] = builder
		}

		return nil
	}
}

func initConfig(ctx *context.Context, cmd *cobra.Command) {
	dir := ctx.WorkingDir

	viper.AddConfigPath(dir)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MIB")
	viper.SetConfigName(Config)
	viper.SetConfigType("yaml")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}

	configPath := viper.GetString(Config)

	if configPath != "" {
		viper.SetConfigFile(configPath)
		configDir := path.Dir(configPath)
		if configDir != "." && configDir != dir {
			viper.AddConfigPath(configDir)
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}

	err := viper.Unmarshal(ctx.Config)
	if err != nil {
		panic(fmt.Errorf("unable to decode into config struct, %v", err))
	}

}
