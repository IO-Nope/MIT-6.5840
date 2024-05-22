package raft

import "log"

// Debugging
const Debug = true
const Debug3A = true
const Debug3B = true

func DPrintf(format string, a ...interface{}) {
	if Debug {
		log.Printf(format, a...)
	}
}
func DPrintfA(format string, a ...interface{}) {
	if Debug3A {
		log.Printf(format, a...)
	}
}
func DPrintfB(format string, a ...interface{}) {
	if Debug3B {
		log.Printf(format, a...)
	}
}
