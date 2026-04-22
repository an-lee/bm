package main

import "testing"

func TestMainDelegatesToExecute(t *testing.T) {
	var called int
	old := executeFn
	executeFn = func() { called++ }
	defer func() { executeFn = old }()
	main()
	if called != 1 {
		t.Fatalf("execute called %d times", called)
	}
}
