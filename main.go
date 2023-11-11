package main

import plugin "kndu/cmd/plugin"

var version string
var commit string

func main() {
	plugin.SetVersion(version)
	plugin.SetCommit(commit)
	plugin.Execute()
}
