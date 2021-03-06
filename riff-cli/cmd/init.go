/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/initializers"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/spf13/cobra"
	"strings"
)

func Init() (*cobra.Command, *options.InitOptions) {

	var initOptions = options.InitOptions{}

	var initCmd = &cobra.Command{
		Use:   "init [language]",
		Short: "Initialize a function",
		Long:  utils.InitCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Initialize(initOptions)
			if err != nil {
				cmd.SilenceUsage = true
				return err
			}
			return nil
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			initOptions.UserAccount = utils.GetUseraccountWithOverride("useraccount", *cmd.Flags())
			if len(args) > 0 {
				if len(args) == 1 && initOptions.FilePath == "" {
					initOptions.FilePath = args[0]
				} else {
					return errors.New(fmt.Sprintf("Invalid argument(s) %v\n", args))
				}
			}

			err := validateInitOptions(&initOptions)
			if err != nil {
				return err
			}
			return nil

		},
	}

	initCmd.PersistentFlags().BoolVar(&initOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	initCmd.PersistentFlags().StringVarP(&initOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	initCmd.PersistentFlags().StringVar(&initOptions.RiffVersion, "riff-version", utils.DefaultValues.RiffVersion, "the version of riff to use when building containers")
	initCmd.PersistentFlags().StringVarP(&initOptions.Version, "version", "v", utils.DefaultValues.Version, "the version of the function image")
	initCmd.PersistentFlags().StringVarP(&initOptions.UserAccount, "useraccount", "u", utils.DefaultValues.UserAccount, "the Docker user account to be used for the image repository")
	initCmd.PersistentFlags().StringVarP(&initOptions.Artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")
	initCmd.PersistentFlags().StringVarP(&initOptions.Input, "input", "i", "", "the name of the input topic (defaults to function name)")
	initCmd.PersistentFlags().StringVarP(&initOptions.Output, "output", "o", "", "the name of the output topic (optional)")
	initCmd.PersistentFlags().BoolVar(&initOptions.Force, "force", utils.DefaultValues.Force, "overwrite existing functions artifacts")

	return initCmd, &initOptions
}

func InitJava(initOptions *options.InitOptions) (*cobra.Command, *options.InitOptions) {

	var initJavaCmd = &cobra.Command{
		Use:   "java",
		Short: "Initialize a Java function",
		Long:  utils.InitJavaCmdLong(),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Java().Initialize(*initOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	initJavaCmd.Flags().StringVar(&initOptions.Handler, "handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	return initJavaCmd, initOptions
}

func InitCommand(initOptions *options.InitOptions) (*cobra.Command, *options.InitOptions) {
	var initCommandCmd = &cobra.Command{
		Use:   "command",
		Short: "Initialize an executable command function",
		Long:  utils.InitCommandCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Command().Initialize(*initOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	return initCommandCmd, initOptions
}

func InitNode(initOptions *options.InitOptions) (*cobra.Command, *options.InitOptions) {
	var initNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Initialize a node.js function",
		Long:  utils.InitNodeCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Node().Initialize(*initOptions)
			if err != nil {
				return err
			}
			return nil
		},
		Aliases: []string{"js"},
	}
	return initNodeCmd, initOptions
}

func InitPython(initOptions *options.InitOptions) (*cobra.Command, *options.InitOptions) {
	var initPythonCmd = &cobra.Command{
		Use:   "python",
		Short: "Initialize a Python function",
		Long:  utils.InitPythonCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			if initOptions.Handler == "" {
				initOptions.Handler = initOptions.FunctionName
			}
			err := initializers.Python().Initialize(*initOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	initPythonCmd.Flags().StringVar(&initOptions.Handler, "handler", "", "the name of the function handler")
	return initPythonCmd, initOptions
}

func InitGo(initOptions *options.InitOptions) (*cobra.Command, *options.InitOptions) {
	var initGoCmd = &cobra.Command{
		Use:   "go",
		Short: "Initialize a go plugin function",
		Long:  utils.InitGoCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			if initOptions.Handler == "" {
				initOptions.Handler = strings.Title(initOptions.FunctionName)
			}
			return initializers.Go().Initialize(*initOptions)
		},
	}
	initGoCmd.Flags().StringVar(&initOptions.Handler, "handler", "", "the name of the function handler (name of Exported go function)")
	return initGoCmd, initOptions
}

func validateInitOptions(options *options.InitOptions) error {
	if err := validateFilepath(&options.FilePath); err != nil {
		return err
	}
	if err := validateFunctionName(&options.FunctionName, options.FilePath); err != nil {
		return err
	}

	if err := validateAndCleanArtifact(&options.Artifact, options.FilePath); err != nil {
		return err
	}

	if err := validateProtocol(&options.Protocol); err != nil {
		return err
	}
	return nil
}
