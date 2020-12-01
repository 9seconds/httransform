// Header and Headers implementation.
//
// You all get used to a fact that headers in HTTP is a map. Actually,
// there is RFC2616 and its friends which tells that:
//
// 1. Header names are case-insensitive.
//
// 2. Header order does not matter.
//
// This is of course not true, and this is a reason why API of
// httransform is so inconvenient at the first glance. The reason is
// that both points are invalid.
//
// Many antibot systems treat header case differently. This is because
// browsers send them differently. Some of them prefer to send all
// headers lowercased. Some of them send headers in 'canonical' form.
// Some of them use mixed approach. This is specific to a browser. And
// if we want to give an ability to replicate that, we have to maintain
// a case of headers.
//
// At the same time, header order really matters. Some headers can be
// merged as comma-delimited lists. Some of them are intentionally
// repeating on a wire like Set-Cookie or just Cookie headers.
//
// A whole reason why httransform is using fasthttp is because this
// library gives an ability to work with headers with no additional
// processing. It does not turn them into map, it does not normalize
// them (it actually does, but there is an option to disable it).
//
// Actually, when you work with headers in httransform, you work with
// instance of Headers struct which maintains a list of headers, not a
// map. But it also gives an ability to get values in case-sensitive
// fashion (and in general case-insensitive).
package headers
