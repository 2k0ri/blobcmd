// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "blobcmd",
	Short: "Microsoft Azure Storage Blob(WASB) command utility",
	Long:  `A tool for accessing Microsoft Azure Storage Blob(WASB).`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// persistent variables
var (
	ConnectionString, AccountName, AccountKey, Container, Prefix, AzureStorageEntrypoint string
	Recursive, DisableHttps                                                              bool
)

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	viper.SetEnvPrefix("AZURE_STORAGE")

	RootCmd.PersistentFlags().StringVarP(&ConnectionString, "connection-string", "c", "", "Storage connection string [AZURE_STORAGE_CONNECTION_STRING]")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.PersistentFlags().StringVarP(&AccountName, "account-name", "a", "", "Storage account name [AZURE_STORAGE_ACCOUNT]")
	viper.RegisterAlias("account", "account-name")
	RootCmd.PersistentFlags().StringVarP(&AccountKey, "account-key", "k", "", "Storage account key [AZURE_STORAGE_ACCESS_KEY]")
	viper.RegisterAlias("access-key", "account-key")
	RootCmd.PersistentFlags().StringVarP(&Container, "container", "C", "", "Storage container name [AZURE_STORAGE_CONTAINER]")
	RootCmd.PersistentFlags().StringVarP(&Prefix, "prefix", "p", "", "Prefix for blob")
	RootCmd.PersistentFlags().BoolVarP(&Recursive, "recursive", "r", false, "Execute recursively")
	RootCmd.PersistentFlags().BoolVar(&DisableHttps, "disable-https", false, "Disable https access")
	RootCmd.PersistentFlags().StringVarP(&AzureStorageEntrypoint, "azure-storage-entrypoint", "E", "core.windows.net", "Azure Storage Entry Point")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blobcmd.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".blobcmd") // name of config file (without extension)
	viper.AddConfigPath("$HOME")    // adding home directory as first search path
	viper.AutomaticEnv()            // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
