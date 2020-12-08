package layers

// Layer is a middleware which processes a request and a response.
//
// You can think about layers as stacks: you go through the list forward
// and on response or error you go backwards. There is a guarantee that
// if you passed through OnRequest function call, your OnResponse is
// also be called.
type Layer interface {
    // OnRequest is going to be executed when your request goes towards
    // an executor.
    //
    // If you return an error from this method, the whole chain is going
    // to be aborted and this error will go backwards via stack.
	OnRequest(*Context) error

    // OnResponse is going to be executed when your response goes from
    // executor or error has happened.
    //
    // If this middleware has generated that error, it will be a first
    // which OnResponse is going to be callled. You need to return an
    // error from this method. Usually you need to return a propagated
    // error but sometimes you can override it adding some context and
    // return a new one.
	OnResponse(*Context, error) error
}
