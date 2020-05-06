package controller

import (
	"github.com/ligangty/api-service/pkg/controller/apiservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, apiservice.Add)
}
