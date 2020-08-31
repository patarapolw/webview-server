package main

import "C"
import "log"

// OnExit execute on exit, including SIGTERM and SIGINT
//export OnExit
func OnExit() C.int {
	log.Println("Executing clean-up function")
	// time.Sleep(2 * time.Second)
	log.Println("Clean-up finished")

	// Cannot seem to use C.void
	return C.int(0)
}
