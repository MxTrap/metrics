package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit direct call in main function"
}

func otherFunc() {
	os.Exit(1) // не должен вызывать ошибку, так как функция не main
}

func mainWithNoExit() {
	println("no exit") // не должен вызывать ошибку
}
