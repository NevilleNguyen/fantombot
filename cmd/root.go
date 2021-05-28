package cmd

import (
	"log"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/quangkeu95/fantom-bot/lib/app"
	"github.com/quangkeu95/fantom-bot/pkg/core"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "start",
	Short: "Fantom bot",
	Long:  "Fantom bot",
	RunE:  rootMain,
}

func rootMain(cmd *cobra.Command, args []string) error {
	coreIns, err := core.New()
	if err != nil {
		return err
	}

	// httpHandler, err := core.NewHttpHandler(coreIns)
	// if err != nil {
	// 	return err
	// }

	// go httpHandler.Run()

	return coreIns.Run()
}

func Execute() {
	l, flush, err := app.NewSugaredLogger()
	if err != nil {
		panic(err)
	}

	defer func() {
		flush()
	}()
	zap.ReplaceGlobals(l.Desugar())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cobra.OnInitialize(config.InitConfig)
	// watch for config file changes

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", "config file (by default app will try to find config in ./env/mainnet.json)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
