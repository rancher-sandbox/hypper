module github.com/rancher-sandbox/hypper

go 1.15

// WARNING! Do NOT replace this without also replacing their lines in the `require` stanza below.
// These `replace` stanzas are IGNORED when this is imported as a library
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

require (
	github.com/Masterminds/log-go v0.4.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/crillab/gophersat v1.3.1
	github.com/fatih/color v1.10.0
	github.com/gofrs/flock v0.8.0
	github.com/gosuri/uitable v0.0.4
	github.com/jinzhu/copier v0.2.8
	github.com/kyokomi/emoji/v2 v2.2.8
	github.com/mattn/go-shellwords v1.0.11
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/thediveo/enumflag v0.10.1
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.6.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.0
	sigs.k8s.io/yaml v1.2.0
)
