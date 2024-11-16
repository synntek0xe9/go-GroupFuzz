package main

import (
	"fmt"
	"net/http"
)

const baseRawReq = `GET / HTTP/1.1
Host: 11
User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Connection: keep-alive
Cookie: wp-settings-time-1=1726671245
Upgrade-Insecure-Requests: 1
Sec-Fetch-Dest: document
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: none
Sec-Fetch-User: ?1

`

type Generator0 struct {
	iter int
}

func NewGenerator0() *Generator0 {
	g := new(Generator0)
	g.iter = 0

	return g
}

func (g *Generator0) next(fuzzData interface{}) (*string, *string, *string, error) {

	count := 110000
	if g.iter >= count {
		return nil, nil, nil, nil
	}
	g.iter = g.iter + 1
	//fmt.Printf("%d\n", g.iter)

	rawReq := baseRawReq
	scheme := "http"
	// var host string
	host := "127.0.0.1:8000"

	return &rawReq, &scheme, &host, nil
}

type Processor0 struct {
}

func (p Processor0) process(resp http.Response, fuzzData interface{}) error {

	fmt.Printf("%s %d\n", resp.Status, resp.ContentLength)
	return nil
}

func main() {

	fmt.Println("ok")
	threadNum := new(int)
	*threadNum = 40
	rate := new(int)
	rate = nil
	err := fuzzAsync(NewGenerator0(), new(Processor0), nil, threadNum, rate)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	return
}
