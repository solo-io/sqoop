// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

package main

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/pkg/errors"
)

// qlooctl/schemaCmd represents the qlooctl/schema command
var schemaCmd = &cobra.Command{
	Use: "schema",
	Aliases: []string{
		"schemas",
	},
	Short: "Create, read, update, and delete GraphQL schemas for QLoo",
	Long:  `Use these commands to register a GraphQL schema with QLoo`,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}

// qlooctl/schemaCmd represents the qlooctl/schema command
var schemaGetCmd = &cobra.Command{
	Use:   "get NAME",
	Short: "return a schema by its name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		msg, err := getSchema(args[0])
		if err != nil {
			return err
		}
		return printAsYaml(msg)
	},
}

func init() {
	schemaCmd.AddCommand(schemaGetCmd)
}

func getSchema(name string) (*v1.Schema, error) {
	cli, err := makeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().Schemas().Get(name)
}
