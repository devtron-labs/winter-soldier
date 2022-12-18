module github.com/devtron-labs/winter-soldier

go 1.13

require (
	github.com/antonmedv/expr v1.9.0
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/go-logr/logr v1.2.3
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	github.com/tidwall/gjson v1.9.3
	//gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	//golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	k8s.io/apiextensions-apiserver v0.25.0
	k8s.io/apimachinery v0.25.0
	k8s.io/client-go v0.25.0
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed
	sigs.k8s.io/controller-runtime v0.13.1
)
