package controller

import (
	"github.com/jcastillo2nd/routed-ip-operator/pkg/controller/clusterroutedipissuer"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusterroutedipissuer.Add)
}
