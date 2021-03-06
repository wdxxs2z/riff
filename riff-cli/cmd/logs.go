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
	"fmt"
	"os/exec"
	"os"
	"bufio"

	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

type LogsOptions struct {
	function  string
	container string
	namespace string
	tail      bool
}

func Logs() *cobra.Command {

	var logsOptions LogsOptions

	// logsCmd represents the logs command
	var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Display the logs for a running function",
		Long: `Display the logs for a running function For example:

    riff logs -n myfunc -t

will tail the logs from the 'sidecar' container for the function 'myfunc'

`,
		Run: func(cmd *cobra.Command, args []string) {

			// get the viper value from env var, config file or flag option
			logsOptions.namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

			fmt.Printf("Displaying logs for container %v of function %v in namespace %v\n\n", logsOptions.container, logsOptions.function, logsOptions.namespace)

			cmdArgs := []string{"get"}
			if logsOptions.namespace != "" {
				cmdArgs = append(cmdArgs, "--namespace", logsOptions.namespace)
			}
			cmdArgs = append(cmdArgs, "pod", "-l", "function=" + logsOptions.function, "-o", "jsonpath={.items[0].metadata.name}")

			output, err := kubectl.ExecForString(cmdArgs)

			if err != nil {
				ioutils.Errorf("Error %v - Function %v may not be currently active\n\n", err, logsOptions.function)
				return
			}

			pod := output

			cmdArgs = []string{"logs"}
			if logsOptions.namespace != "" {
				cmdArgs = append(cmdArgs, "--namespace", logsOptions.namespace)
			}
			cmdArgs = append(cmdArgs, pod, "-c", logsOptions.container)
			if logsOptions.tail {
				cmdArgs = append(cmdArgs, "-f")

				kubectlCmd := exec.Command("kubectl", cmdArgs...)
				cmdReader, err := kubectlCmd.StdoutPipe()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for kubectlCmd", err)
					return
				}

				scanner := bufio.NewScanner(cmdReader)
				go func() {
					for scanner.Scan() {
						fmt.Printf("%s\n\n", scanner.Text())
					}
				}()

				err = kubectlCmd.Start()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error starting kubectlCmd", err)
					return
				}

				err = kubectlCmd.Wait()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error waiting for kubectlCmd", err)
					return
				}

			} else {

				output, err := kubectl.ExecForString(cmdArgs)

				if err != nil {
					ioutils.Errorf("Error: %v\n", err)
					return
				}

				fmt.Printf("%v\n", output)
			}

		},
	}
	logsCmd.Flags().StringVarP(&logsOptions.function, "name", "n", "", "the name of the function")
	logsCmd.Flags().StringVarP(&logsOptions.container, "container", "c", "sidecar", "the name of the function container (sidecar or main)")
	logsCmd.Flags().BoolVarP(&logsOptions.tail, "tail", "t", false, "tail the logs")

	logsCmd.Flags().StringP("namespace", "", "", "the namespace used for the deployed resources")

	logsCmd.MarkFlagRequired("name")
	return logsCmd
}
