// A wrapper for LRU/LFU cache used by httransform
//
// The thing is that there are many (good) caching libraries available
// in Go. So many so I even think to make my own. But the truth is that
// I feel a huge discomfort leaking implemetation details in the rest
// of the code base. When I add a parameter which is unique only for
// specific implementation and we have to support it until end of the
// life.
//
// So, the idea of this package is to introduce a minimal interface I'm
// going to use for httransform.
package cache
