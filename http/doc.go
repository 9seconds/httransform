// A custom implementation of HTTP request sender.
//
// A custom implementation is required because fasthttp's one does not
// support streaming response bodies. So, if you start to use proxies to
// download gigabytes of data, you are going to have serious problems.
package http
