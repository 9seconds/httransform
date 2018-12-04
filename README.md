httransform
===========

[![Build Status](https://travis-ci.org/9seconds/httransform.svg?branch=master)](https://travis-ci.org/9seconds/httransform)
[![CodeCov](https://codecov.io/gh/9seconds/httransform/branch/master/graph/badge.svg)](https://codecov.io/gh/9seconds/httransform)
[![GoDoc Reference](https://camo.githubusercontent.com/7540274b4c20318e1b1f2d8abe11ba136c926233/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f76616c79616c612f66617374687474703f7374617475732e737667)](https://godoc.org/github.com/9seconds/httransform)
[![Go Report Card](https://goreportcard.com/badge/github.com/9seconds/httransform)](https://goreportcard.com/report/github.com/9seconds/httransform)

httransform is the library/framework to build your own HTTP
proxies. It relies on high-performant and memory-efficient
[fasthttp](https://github.com/valyala/fasthttp) library as HTTP base
layer and can give you an ability to build a proxy where you control an
every aspect.

Main features of this framework:

1. Support of HTTP (plain HTTP) proxy protocol.
2. Support of HTTPS (with CONNECT method) protocol. This library does MITM
   and provides a possibility to generate TLS certificates for the hosts
   on-the-fly.
3. Keeps and maintans the order of header and their case (no normalization).
4. Support the concept of _layers_ or middlewares which process HTTP
   requests and responses
5. Supports custom _executors_: a functions which converts HTTP request to
   HTTP response. So, your proxy can fetch the data from other services,
   which are not necessary HTTP. Executor simply converts HTTP request
   structure to HTTP response.

Please check [full
documentation](https://godoc.org/github.com/9seconds/httransform) to get
more details.
