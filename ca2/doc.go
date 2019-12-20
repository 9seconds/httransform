// Package ca contains functions to manage lifecycle of TLS CA.
//
// This CA is required to generate TLS certificate for hostnames on the
// fly. It uses self-signed certificate + its primary key (or, if you
// want, you can provide your own certificates) to generate ad-hoc TLS
// certificates for the given hosts.
//
// The certificates are generated in determenistic way derived from your
// CA private key so please keep it is secret.
//
// To generate your own set of CA certificate and private key, please
// use the following command line:
//
//   openssl req -x509 -newkey rsa:1024 -keyout private-key.pem -out ca.crt -days 3650 -nodes
//
// file ca.crt will contain CA certificate; private-key.pem - CA private
// key.
package ca2
