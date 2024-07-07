package common

import (
	"os"
	"strings"
)

// When on the prometheus pod, the following environment variables are set.
// This means we can make educated guesses on how to connect to the k8s API
const (
	EnvVarHostname = "HOSTNAME"
	PodNamePartial = "cloudzero-agent-server"
)

func InPod() bool {
	value := os.Getenv(EnvVarHostname)
	return strings.Contains(value, PodNamePartial)
}
