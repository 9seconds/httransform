// Default implementation of executor.
//
// In httransform Executor is something which does end processing of the
// given Context instance. It usually takes an HTTP request and fills
// HTTP response. It actually has to do that but in reality it can do
// much more, like request hijacking to support websockets or plain TCP
// upgrades.
//
// So, Executor is a function which transforms HTTP request to HTTP
// response. And returns error if something goes wrong during that
// process.
package executor
