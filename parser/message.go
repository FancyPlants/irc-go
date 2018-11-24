package parser

// Message is a parsed representation of a message
// from an IRC server
type Message struct {
	// TODO: make this a map[string]string at some point or other
	Tags       string
	Source     string
	Command    string
	Parameters []string
}
