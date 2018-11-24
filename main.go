package main

import (
	"fmt"
	"bufio"
	"os"

	"github.com/fatih/color"
)

var green = color.New(color.FgGreen)

func main() {
	client := NewIRC("irc.freenode.net:6667", true)
	msgs := client.WatchMessages()
	go func() {
		for {
			msg := <-msgs

			fmt.Printf("%+v\n", msg)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		writer.WriteString(green.Sprint("Client > "))
		writer.Flush()
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		switch (input) {
		case "login\n":
			client.Login("fancyplants", "Joshua Flancer", "fancyplants")
		case "join\n":
			client.JoinChannel("#node")
		case "part\n":
			client.LeaveChannel("#node")

		default:
			fmt.Println("Unknown.")
		}
	}
}
