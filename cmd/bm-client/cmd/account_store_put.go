// Copyright (c) 2021 BitMaelum Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/internal"
	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/internal/api"
	"github.com/bitmaelum/bitmaelum-suite/internal/vault"
	"github.com/spf13/cobra"
)

var accountStorePutCmd = &cobra.Command{
	Use:   "put",
	Short: "Display store contents",
	Run: func(cmd *cobra.Command, args []string) {
		v := vault.OpenDefaultVault()

		info, err := vault.GetAccount(v, *astAccount)
		if err != nil {
			fmt.Println("cannot find account in vault")
			os.Exit(1)
		}

		resolver := container.Instance.GetResolveService()
		routingInfo, err := resolver.ResolveRouting(info.RoutingID)
		if err != nil {
			fmt.Println("cannot resolve routing")
			os.Exit(1)
		}

		client, err := api.NewAuthenticated(*info.Address, info.GetActiveKey().PrivKey, routingInfo.Routing, internal.JwtErrorFunc)
		if err != nil {
			fmt.Println("cannot connect to API")
			os.Exit(1)
		}

		err = client.StorePutValue(info.Address.Hash(), *aspKey, *aspValue)

		// table := tablewriter.NewWriter(os.Stdout)
		// table.SetHeader([]string{"Key", "Value"})
		//
		// table.Append([]string{"Name", info.Name})
		//
		// if info.Settings != nil {
		// 	for k, v := range info.Settings {
		// 		table.Append([]string{k, v})
		// 	}
		// }
		//
		// table.Render()
	},
}

var (
	aspKey *string
	aspValue *string
)

func init() {
	accountStoreCmd.AddCommand(accountStorePutCmd)

	aspKey = accountStorePutCmd.PersistentFlags().String("key", "", "Key to store")
	aspValue = accountStorePutCmd.PersistentFlags().String("value", "", "Value to store")

	_ = accountStorePutCmd.MarkPersistentFlagRequired("key")
	_ = accountStorePutCmd.MarkPersistentFlagRequired("value")
}
