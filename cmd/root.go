/*
Copyright © 2024 George <george@betterde.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/betterde/cdns/internal/journal"
	"github.com/betterde/cdns/internal/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	app     *fiber.App
	name    = "CDNS"
	build   = "current"
	commit  = "none"
	version = "develop"
	verbose bool
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cdns",
	Short:   "An open-source lightweight DNS server that switches to ACME DNS challenge.",
	Version: fmt.Sprintf("Version: %s\nBuild at: %s\nCommit hash: %s", version, build, commit),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	app = fiber.New(fiber.Config{
		AppName:       name,
		ServerHeader:  fmt.Sprintf("%s %s", name, rootCmd.Version),
		CaseSensitive: true,
		// Override default error handler
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a fiber.*Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			if err != nil {
				if code >= fiber.StatusInternalServerError {
					journal.Logger.Errorw("Analysis server runtime error:", zap.Error(err))
				}

				// In case the SendFile fails
				return ctx.Status(code).JSON(response.Send(code, err.Error(), err))
			}

			return nil
		},
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .cdns.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose mode")

	rootCmd.SetVersionTemplate("{{printf \"%s\" .Version}}\n")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	journal.InitLogger()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".orbit" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName(".cdns")
		viper.AddConfigPath("/etc/cdns")
	}

	// read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CDNS")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		journal.Logger.Errorf("Failed to read configuration file: %s", err)
		os.Exit(1)
	}

	viper.AutomaticEnv() // read in environment variables that match

	level := viper.GetString("logging.level")
	if verbose {
		level = "DEBUG"
	}

	err := journal.SetLevel(level)
	if err != nil {
		journal.Logger.Error("Unable to set logger level", err)
		os.Exit(1)
	}

	journal.Logger.Debugf("Configuration file currently in use: %s", viper.ConfigFileUsed())
}
