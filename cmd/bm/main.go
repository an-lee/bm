package main

import "bm/internal/cli"

// executeFn is swapped in tests so main() stays covered without running the real CLI.
var executeFn = cli.Execute

func main() {
	executeFn()
}
