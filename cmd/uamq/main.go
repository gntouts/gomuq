package main

var uartOut = make(chan string, 1)
var mqttOut = make(chan string, 1)

func main() {
	app := app()
	app.start()
}
