/*
Copyright © 2023 George

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
	"fmt"
	"github.com/betterde/cdns/internal/journal"
	"github.com/betterde/cdns/pkg/api"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// routeCmd represents the route command
var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "List all registered routes",
	Run: func(cmd *cobra.Command, args []string) {
		routes := api.ServerInstance.Engine.GetRoutes(true)
		writer := tabwriter.NewWriter(os.Stdout, 40, 0, 0, '.', tabwriter.TabIndent|tabwriter.Debug)

		for _, route := range routes {
			if route.Method == "HEAD" {
				continue
			} else if route.Method == "GET" {
				route.Method = "GET|HEAD"
			}

			if route.Name == "" {
				route.Name = "."
			}

			if _, err := fmt.Fprintln(writer, route.Method, "\t", route.Path, "\t", route.Name, "\t"); err != nil {
				journal.Logger.Sugar().Error(err)
				os.Exit(1)
			}
		}

		if err := writer.Flush(); err != nil {
			journal.Logger.Sugar().Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(routeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// routeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// routeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
