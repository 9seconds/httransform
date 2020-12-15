// Different connection upgraders.
//
// Usually you work in request/response mode: send HTTP request,
// get a response back. But web is much harder nowadays: there is a
// possibility to perform a connection upgrade.
//
// Connection upgrade means that after webserver respond with success,
// a connection is considered as plain TCP one. You can pass bytes here
// and there, you can perform nested TLS connections etc. Usually people
// upgrade their connection to support websockets but connection upgrade
// is more general thing than just websockets.
//
// This package contains 2 main implementation you probably need: TCP
// and websockets. Both implementations are read-only, you cannot alter
// a content. But if you want, you can use them to build your own
// implementations. Both of them are simple enough.
package upgrades
