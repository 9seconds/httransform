// Context and layers.
//
// A layer is a keystone of extensibility of httransform.
// You can think about them as stacks of callbacks.
//
// One diagram worth 1000 words:
//
//          HTTP interface            Layer 1             Layer 2
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
// backward. This is a contract of this package. So, if you have a
// layers A and B, a request path is A(onRequest) -> B(onRequest) ->
// Executor -> B(onResponse) -> A(onResponse) -> client. So, that's
// why it worth to think about layers as about 'stacks'.
//
// Let's check what happens if some layer returns an error:
//
//          HTTP interface            Layer 1                 Layer 2
//        +----------------+      **************           **************
//        |                |      *            *           *            *       ==============
//   ---> |  HTTP request  | ===> *  OnRequest * ===> X    *  OnRequest *       =            =
//        |                |      *            *      |    *            *       =            =
//        +----------------+      **************      |    **************       =  Executor  =
//        |                |      *            *      |    *            *       =            =
//   <--- |  HTTP response | <=== * OnResponse * <=== +    * OnResponse *       =            =
//        |                |      *            *           *            *       ==============
//        +----------------+      **************           **************
//
// So, we guarantee that error will pass through the same layers which
// were already processed a request.
//
// Executor can also return an error. But it is its responsibility to
// put this error into a response.
package layers
