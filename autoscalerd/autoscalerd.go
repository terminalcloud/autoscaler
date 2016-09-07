package main

import (
	"github.com/terminalcloud/autoscaler"
)

func main() {
	autoscaler.Configure()
	autoscaler.StartAutoscaler()
}
