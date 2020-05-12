package main

import (
	"bitbucket.it.keysight.com/qsr/clipsgo.git/pkg/clips"
)

func main() {
	env := clips.CreateEnvironment()
	env.Shell()
}
