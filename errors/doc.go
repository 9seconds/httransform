// Custom errors.
//
// Actually, common errors are also good but we can do better. And we
// need to do better if we want to unleash a potential to return great
// errors to clients. This package extends a standard library and adds
// its own wrappers we encourage to use. The main benefit of doing that
// is consistency in error management, integration into our JSONified
// error reporting etc.
//
// When to use this errors? We encourage you to use Errors when you work
// with layers or executors. You can of course return any errors you
// like but then httransform is going to be limited in how it processes
// a data.
package errors
