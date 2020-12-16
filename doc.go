// httransform is a framework to build a various MITM proxies.
//
// This framework is built using fasthttp
// (https://github.com/valyala/fasthttp) library to provide fast
// processing of HTTP requests. Also, as fasthttp it actively
// uses object pooling to reduce the pressure on GC and possibly
// save a little bit of memory. Unlike its alternative, goproxy
// (https://github.com/elazarl/goproxy), httransform does not have
// disadvantages of using net/http: we can keep header case and order.
//
// This framework provides features to build robust high-performant HTTP
// and HTTPS proxies (with the support of the CONNECT method and TLS
// certificate generation on-the-fly), it supports middleware layers and
// custom executors. It is also able to correctly process different http
// connection upgrades like websocket support.
//
// Layers are middlewares which do preprocess of the requests or
// postprocessing of the response. The thing about layers is that you
// have to define both directions: to executor and from executor. We
// encourage to create layers which know as less as possible about each
// other. If, for example, you raise an error in one layer, this error
// has to be processed only there, other layers should ignore that. This
// framework does not restrict any deviations from this rule, but you'll
// get more simple and clean design if you treat layers as independent
// as possible.
//
// An executor is some function which converts HTTP request to response.
// In the most simple case, an executor is HTTP client. But also, you
// can convert the request to JSON, send it somewhere, get ProtoBuf back
// and convert it to HTTP response. Or you can play with NATS. Or 0mq.
// Or RabbitMQ. You got the idea. Executor defines the function which
// converts HTTP request to HTTP response. If connection upgrade is
// required, this is a responsibility of the executor to make it.
//
// If any of your layers creates an error, executor won't be called.
// HTTP response will be converted to 500 response and error would be
// propagated by the chain of layers back.
//
// A Mental Scheme of the Request
//
// Just take a look here:
//
//           HTTP interface            Layer 1             Layer 2
//        +----------------+      **************      **************
//        |                |      *            *      *            *       ==============
//   ---> |  HTTP request  | ===> *  OnRequest * ===> *  OnRequest * ===>  =            =
//        |                |      *            *      *            *       =            =
//        +----------------+      **************      **************       =  Executor  =
//        |                |      *            *      *            *       =            =
//   <--- |  HTTP response | <=== * OnResponse * <=== * OnResponse * <===  =            =
//        |                |      *            *      *            *       ==============
//        +----------------+      **************      **************
//
// As you see, the request goes through the all layers forward and
// backward. This is a contract of this package.
//
// Features
//
// 1. Interface of HTTP proxy
//
// 2. Interface of HTTPS proxy (e.g CONNECT method)
//
// 3. CONNECT method can be used to establish plain TCP tunnels. Upgrade
// to TLS is not mandatory.
//
// 4. Both HTTP request and responses can be processed: headers
// are added/removed/changed and so on.
//
// 5. TCP connection upgrade is supported. Websockets are supported. But
// by default you can't interfere: you can just watch.
package httransform
