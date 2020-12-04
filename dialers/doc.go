// Implementation of different dialers.
//
// You can think about dialers as implementation of net.Dialer on
// steroids. The goal of each dialer is to return an established
// net.Conn with a target netloc. If it required to use proxies, this is
// not a responsibility of the Transport or any other concepts, this is
// a responsibility of Dialer.
//
// This might sound confusing for you but if you think that we have to
// support websockets and generic upgrades as well as HTTP requests, it
// is quite appealing to delegate these operation to dialer.
//
// For example, if you work with http proxies, a goal of dialer is
// to make all these ceremonies like CONNECT tunneling, TLS upgrade
// whatsoever.
//
// Another important consideration is that dialer has to do 3 different
// things:
//
// 1. To Dial, establish a TCP connection with a target netloc. A result
// is the socket which talks with a target.
//
// 2. To upgrade a socket from the step 1 to TLS. Sometimes it is
// trivial but in case of HTTP proxies, for example, we have to make an
// additional steps like CONNECT ceremony.
//
// 3. Patch HTTP request. Sometimes we want to process a request in a
// special way for this socket. This method prepares an HTTP request so
// it can be passed down to a socket.
package dialers
