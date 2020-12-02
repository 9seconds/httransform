// Event stream implementation.
//
// You may get used to idea of passing many different interfaces to
// collect metrics, track some activity, emit logs etc. To stop bloating
// interfaces, httransform chooses another approach: it implements event
// stream.
//
// Basic idea is dumb and simple: each time when something interesting
// is happening, httransform or users of the library send special
// events. These events can be processed and used to identify start/stop
// of the request processing, failures etc.
//
// Actually, you can extend this framework and use your own values and
// process them in the same fashion.
//
//     import "github.com/github.com/9seconds/httransform/v2/events"
//
//     const (
//         MyEvent      events.EventType = events.EventTypeUserBase + iota
//         AnotherEvent
//     )
//
// If you do that, I recommend you to keep these constants in a single
// module so you can easily control a uniqueness of values.
package events
