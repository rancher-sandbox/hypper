package action

import (
	"fmt"

	helmAction "helm.sh/helm/v3/pkg/action"
)

func RunInstall(action *helmAction.Configuration, client *helmAction.Install, args []string) error {

	fmt.Println("there is no logic yet")
	fmt.Println(args[0])
	return nil
}
