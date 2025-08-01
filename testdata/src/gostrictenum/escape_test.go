// nolint
package main

import (
	"context"
	"fmt"
	"net/http"
)

// SOAP response types
type soapResponseType int

const (
	soapResponseOk soapResponseType = iota
	soapResponseNotSOAP
	soapResponseError
	soapResponseRedirected
	httpErrorCannotConnect
)

type soapResponseTypeInvalid int

const (
	soapResponseTypeInvalidA soapResponseTypeInvalid = iota
	soapResponseTypeInvalidB
	soapResponseTypeInvalidC
)

type Evidence struct{}

func IsSOAPService(ctx context.Context, client http.Client, url string) (soapResponseType, *Evidence) {
	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Invalid URL", err)
		return soapResponseNotSOAP, nil
	}
	// ...

	return soapResponseOk, nil
}

func IsSOAPServiceInvalid(ctx context.Context, client http.Client, url string) (soapResponseType, *Evidence) {
	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Invalid URL", err)
		return soapResponseNotSOAP, nil
	}
	// ...

	return 1, nil // want "returning literal '1' which is not a valid enum value for type soapResponseType"
}

func IsSOAPServiceInvalid2(ctx context.Context, client http.Client, url string) (soapResponseTypeInvalid, *Evidence) {
	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Invalid URL", err)
		return soapResponseTypeInvalidA, nil
	}
	// ...

	return soapResponseTypeInvalidA, nil
}
