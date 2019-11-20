package controller

import (
	"github.com/vfreex/release-engine-prototype/pkg/controller/source"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, source.Add)
}
