package apiservice

import (
	openshiftutils "github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"k8s.io/client-go/rest"
)

// isOpenShift returns true when the container platform is detected as OpenShift
func isOpenShift(c *rest.Config) bool {
	isOpenShift, err := openshiftutils.IsOpenShift(c)
	if err != nil {
		return false
	}
	return isOpenShift
}
