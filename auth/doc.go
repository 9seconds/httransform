// Authenticators interface and implementations.
//
// A goal of authenicator is to identify how to extract a user from the
// request and checks if it is possible to proceed next. As you see,
// we intentionally mix authentication and authorization here because
// everyone does it. And if we start to distringuish these 2 different
// concepts, people will immediately start to be confused.
//
// So that's why a package name is auth. We do auth here.
package auth
