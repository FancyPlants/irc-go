package main

import (
	"fmt"
	"net"
	"bufio"

	"irc-go/parser"
	"github.com/fatih/color"
)

/*
	* At the end of the day, all computer science boils down to is
	* taking some data and transforming it into other data. LETS DO THIS
*/

var bgRed = color.New(color.BgRed).Add(color.Underline)

// IRC represents a simple IRC client
type IRC struct {
	Nickname string
	Fullname string
	Username string

	conn net.Conn
	connWriter *bufio.Writer
	connReader *bufio.Reader
	p *parser.Parser


	// Since I miss the concept of events from NodeJS, here
	// is my best emulation of that using channels. There will (hopefully)
	// be an external system that takes a single channel and resends it
	// to any registered listeners
	errChannel chan error
	msgChannel chan parser.Message

	// if anything is sent, then boom, kill this connection
	stopChannel chan bool
}

// NewIRC exists in case of possible initialization needed for IRC client
func NewIRC(address string, debug bool) *IRC {
	errChannel := make(chan error, 50)
	msgChannel := make(chan parser.Message, 50)

	conn, err := net.Dial("tcp4", address)
	if err != nil {
		errChannel <- err
	}

	irc := &IRC {
		conn: conn,
		errChannel: errChannel,
		msgChannel: msgChannel,
		p: parser.NewParser(msgChannel),
	}

	go irc.watchStop()

	if debug {
		go irc.reportErrors()
	}

	irc.connReader = bufio.NewReader(conn)
	go irc.receiveData()
	irc.connWriter = bufio.NewWriter(conn)

	return irc
}

// * Private functions

func (irc *IRC) handleErr(err error) {
	if err != nil {
		irc.errChannel <- err
	}
}

// reportErrors should be run in a goroutine.
func (irc *IRC) reportErrors() {
	for {
		select {
		default:
			bgRed.Printf("%s\n", <-irc.errChannel)

		case <-irc.stopChannel:
			return
		}
		
	}
}

func (irc *IRC) receiveData() {
	for {
		select {
		default:
			input, err := irc.connReader.ReadString('\n')
			irc.handleErr(err)
			irc.p.ParseChunk(input)

		case <-irc.stopChannel:
			return
		}
	}
}

func (irc *IRC) watchStop() {
	select {
	case <-irc.stopChannel:
		irc.conn.Close()
	}
}

// * Public functions

// ** Utilities

// WatchMessages returns a read-only version of the internal
// message channel to ensure no outside tampering
func (irc *IRC) WatchMessages() <-chan parser.Message {
	return irc.msgChannel
}

// WatchErrors returns a read-only channel to make sure
// errors can be watched for externally
func (irc *IRC) WatchErrors() <-chan error {
	return irc.errChannel
}

// ** Actions

// Login sends the NICK message followed by USER.
// hopefully one day I can implement passwords n stuff
func (irc *IRC) Login(nickname, fullname, username string) {
	irc.Nickname = nickname
	irc.Fullname = fullname
	irc.Username = username

	_, err := fmt.Fprintf(irc.connWriter, "NICK %s\r\n", nickname)
	irc.handleErr(err)

	_, err = fmt.Fprintf(irc.connWriter, "USER %s 0 * :%s\r\n", username, fullname)
	irc.handleErr(err)

	irc.connWriter.Flush()
}

// JoinChannel sends the JOIN message to get into a channel
func (irc *IRC) JoinChannel(channel string) {
	_, err := fmt.Fprintf(irc.connWriter, "JOIN %s\r\n", channel)
	irc.handleErr(err)
	irc.connWriter.Flush()
}

// LeaveChannel sends the PART message to leave a channel
func (irc *IRC) LeaveChannel(channel string) {
	_, err := fmt.Fprintf(irc.connWriter, "PART %s\r\n", channel)
	irc.handleErr(err)
	irc.connWriter.Flush()
}
