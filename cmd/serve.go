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
	"context"
	"github.com/betterde/cdns/global"
	"github.com/betterde/cdns/internal/journal"
	"github.com/betterde/cdns/pkg/api"
	"github.com/betterde/cdns/pkg/dns"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start an orbit server",
	Run: func(cmd *cobra.Command, args []string) {
		errChan := make(chan error, 1)
		// Run domain name server
		dns.InitServer(errChan)

		// Run HTTP server
		api.ServerInstance.Run(verbose, errChan)

		if err := signalHandler(errChan, global.CancelFunc); err != nil {
			journal.Logger.Sugar().Errorw("Failed to shutdown CDNS server:", err)
		}
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
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func signalHandler(errChan chan error, cancel context.CancelFunc) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case <-shutdown:
		cancel()
		for _, server := range dns.Servers {
			err := server.Server.Shutdown()
			if err != nil {
				journal.Logger.Sugar().Error("Failed to shutdown server:", err)
			}
		}
		return api.ServerInstance.Engine.Shutdown()
	case err := <-errChan:
		cancel()
		journal.Logger.Sugar().Panic(err)
		return api.ServerInstance.Engine.Shutdown()
	}
}
