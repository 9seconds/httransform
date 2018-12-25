// Package client contains a simple implementation of HTTP 1 client
// which returns a streaming body.
//
// By default fasthttp puts whole response into the memory. This is fine
// for small responses but if we use this library to stream huge files,
// this can cause a lot of problems.
//
// WHen fasthttp will introduce support of streaming bodies, this
// package should be considered as obsolete.
package client
