package internal

import "sync"

// Placeholder is a placeholder object that can be used globally.
var Placeholder PlaceholderType

type (
	// GenericType can be used to hold any type.
	GenericType = interface{}
	// PlaceholderType represents a placeholder type.
	PlaceholderType = struct{}
)

// A DoneChan is used as a channel that can be closed multiple times and wait for done.
type DoneChan struct {
	done chan PlaceholderType
	once sync.Once
}

// NewDoneChan returns a DoneChan.
func NewDoneChan() *DoneChan {
	return &DoneChan{
		done: make(chan PlaceholderType),
	}
}

// Close closes dc, it's safe to close more than once.
func (dc *DoneChan) Close() {
	dc.once.Do(func() {
		close(dc.done)
	})
}

// Done returns a channel that can be notified on dc closed.
func (dc *DoneChan) Done() chan PlaceholderType {
	return dc.done
}
