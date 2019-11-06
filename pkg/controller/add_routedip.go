package controller

import (
	"github.com/jcastillo2nd/routedip-operator/pkg/controller/routedip"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, routedip.Add)
}
