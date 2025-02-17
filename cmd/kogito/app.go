package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

var appCmd *cobra.Command
var appCmdName string

var _ = RegisterCommandVar(func() {
	appCmd = &cobra.Command{
		Use:   "app NAME",
		Short: "Sets the Kogito application where your Kogito Service will be deployed",
		Long:  `app will set the context where the Kogito services will be deployed. It's the namespace/project on Kubernetes/OpenShift world.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(appCmd)
	appCmd.Flags().StringVarP(&appCmdName, "name", "n", "", "The app name")
})

func appExec(cmd *cobra.Command, args []string) error {
	if len(newAppName) == 0 {
		if len(args) == 0 {
			if len(config.Namespace) == 0 {
				return fmt.Errorf("No application set in the context. Use 'kogito new-app NAME' to create a new application")
			}
			log.Infof("Application in the context is '%s'. Use 'kogito deploy NAME SOURCE' to deploy a new application.", config.Namespace)
			return nil
		}
		appCmdName = args[0]
	}

	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(appCmdName); err != nil {
		return fmt.Errorf("Error while trying to look for the namespace. Are you logged in? %s", err)
	} else if ns != nil {
		config.Namespace = appCmdName
		log.Infof("Application set to '%s'", appCmdName)
		return nil
	}

	return fmt.Errorf("Application '%s' not found. Try running 'kogito new-app %s' to create your application first", appCmdName, appCmdName)
}
