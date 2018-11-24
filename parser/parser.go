package parser

import (
	"strings"

	"github.com/fatih/color"
)

// FOR CONSOLE DEBUGGING
var bgRed = color.New(color.BgRed).Add(color.Underline)

// Parser is an instance of chunk processor. How v1 of 
// Parser works is, on creation, it takes in a writable channel
// and every successfully constructed message is emitted through it.
type Parser struct {
	msgBuffer string
	output    chan<- Message
	msgChannel chan string
}

// NewParser returns a new instance of message parser
func NewParser(output chan<- Message) *Parser {
	p := &Parser {
		msgBuffer: "",
		output: output,
		msgChannel: make(chan string, 100),
	}

	go p.parseString()

	return p
}

// ParseChunk takes a chunk of bytes from a socket probably
// and then tries to piece them together and then send them
// off to another channel
func (p *Parser) ParseChunk(chunk []byte) {
	strChunk := string(chunk)
	strs := strings.SplitAfter(strChunk, "\r\n")

	for _, message := range strs {
		var finalMsg string
		if p.msgBuffer != "" {
			finalMsg = message + p.msgBuffer
			bgRed.Printf("Msgbuffer used: '%s' + '%s'\n", finalMsg, p.msgBuffer)
			p.msgBuffer = ""
		} else if !strings.HasSuffix(message, "\r\n") {
			p.msgBuffer = message
			continue
		}

		p.msgChannel <- finalMsg
	}
}

// parseString should be run in a goroutine upon a parser's creation
// and keeps taking messages out of the message channel and sends it into the
// output channel
func (p *Parser) parseString() {
	for {
		msg := <-p.msgChannel
		msgStruct := Message {}

		parts := strings.Split(msg, " ")

		commandParsed := false

		for index, part := range parts {
			if index == 0 && part[0] == '@' { // tags
				msgStruct.Tags = part
			} else if (index == 0 || index == 1) && part[0] == ':' { // sources/prefixes
				msgStruct.Source = part
			} else if part[0] == ':' { // final parameter marked by ':'
				// TODO: possibly turn into string builder for ULTRA SPEEDUP
				finalParam := ""
				for i := index; i < len(parts); i++ {
					finalParam += parts[i]
				}
				msgStruct.Parameters = append(msgStruct.Parameters, finalParam)
				break
			} else { // either a command or normal param
				if !commandParsed {
					msgStruct.Command = part
					commandParsed = true
				} else {
					msgStruct.Parameters = append(msgStruct.Parameters, part)
				}
			}
		}

		p.output <- msgStruct
	}
}
