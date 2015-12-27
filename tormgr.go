package main

func main() {

	if err := InitServices(); err != nil {
		panic(err)
	}
	setupServer()
}
