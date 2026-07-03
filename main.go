package main

import "github.com/yourusername/argus/cmd"

// main is intentionally tiny — all logic lives in cmd/
// This is the standard Cobra project layout
func main() {
	cmd.Execute()
}
