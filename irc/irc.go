package irc

import (
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
			data := make([]byte, 4096)
			_, err := irc.connReader.Read(data)
			if err != nil {
				irc.errChannel <- err
			}

			irc.p.ParseChunk(data)

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

// Login sends the NICK message followed by USER.
// hopefully one day I can implement passwords n stuff
func (irc *IRC) Login(nickname, fullname, username string) {

}