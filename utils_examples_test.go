package httransform

import (
	"encoding/base64"
	"fmt"
	"net/url"
)

func ExampleExtractAuthentication() {
	username := "user"
	password := "password"
	encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	user, pass, _ := ExtractAuthentication([]byte(encoded))

	fmt.Println(encoded)
	fmt.Println(string(user))
	fmt.Println(string(pass))
	// Output:
	// Basic dXNlcjpwYXNzd29yZA==
	// user
	// password
}

// This exqmple shows how to use this function with both user and password.
func ExampleMakeProxyAuthorizationHeaderValue_full() {
	u := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
		User:   url.UserPassword("user", "password"),
	}

	fmt.Println(u)
	fmt.Println(string(MakeProxyAuthorizationHeaderValue(u)))
	// Output:
	// http://user:password@127.0.0.1:3128
	// Basic dXNlcjpwYXNzd29yZA==
}

// This example shows how to use this function if you have only user.
func ExampleMakeProxyAuthorizationHeaderValue_user() {
	u := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
		User:   url.User("user"),
	}

	fmt.Println(u)
	fmt.Println(string(MakeProxyAuthorizationHeaderValue(u)))
	// Output:
	// http://user@127.0.0.1:3128
	// Basic dXNlcjo=
}

// This example shows how this function behaves if we provide no credentials.
func ExampleMakeProxyAuthorizationHeaderValue_nothing() {
	u := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}

	fmt.Println(u)
	fmt.Println(MakeProxyAuthorizationHeaderValue(u))
	// Output:
	// http://127.0.0.1:3128
	// []
}
