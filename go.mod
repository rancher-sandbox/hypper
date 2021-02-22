module github.com/rancher-sandbox/hypper

go 1.15

require (
	github.com/Masterminds/log-go v0.4.0
	github.com/containerd/containerd v1.4.3
	github.com/deislabs/oras v0.10.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-units v0.4.0
	github.com/fatih/color v1.10.0
	github.com/gosuri/uitable v0.0.4
	github.com/kyokomi/emoji/v2 v2.2.8
	github.com/mattn/go-shellwords v1.0.10
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/cli-runtime v0.20.4
)

replace (
	// WARNING! Do NOT replace these without also replacing their lines in the `require` stanza below.
	// These `replace` stanzas are IGNORED when this is imported as a library
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/mattfarina/hypper => github.com/rancher-sandbox/hypper v0.0.0-20210218091504-4d1470834315
)
