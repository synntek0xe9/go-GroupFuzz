package main

import (
	"bufio"
	"container/ring"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"time"
)

func makeHttpReq(raw string, scheme *string, host *string) (*http.Request, error) {
	if scheme == nil {
		*scheme = "http"
	}
	r, err := http.ReadRequest(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		return nil, err
	}
	r.RequestURI, r.URL.Scheme, r.URL.Host = "", *scheme, r.Host
	if host != nil {
		r.URL.Host = *host
	}
	return r, nil
}

type FuzzArgGenerator interface {
	next(interface{}) (*string, *string, *string, error)
}

type FuzzPostProcessor interface {
	process(http.Response, interface{}) error
}

func fuzz(generator FuzzArgGenerator, postProcessor FuzzPostProcessor, data interface{}) error {

	rawReq, scheme, host, err := generator.next(data)

	if err != nil {
		return err
	}

	for rawReq != nil {
		if scheme == nil {
			return errors.New("rawReq exists, even tho scheme doesn't exisit") // TODO: warning instead?
		}

		httpReq, err := makeHttpReq(*rawReq, scheme, host)
		if err != nil {
			return err // TODO: warning instead
		}

		transport := http.Transport{DisableKeepAlives: false, DisableCompression: true} // disable all parameters that modifies request (can't edit "Proxy-Authorization" handling)

		client := http.Client{}

		client.Transport = &transport

		httpResp, err := client.Do(httpReq)

		if err != nil {
			return err // TODO: warning instead
		}

		if postProcessor != nil {

			if httpResp == nil {
				return errors.New("httpResp is null") // TODO: warning instead
			} else {
				err = postProcessor.process(*httpResp, data)
				if err != nil {
					return errors.New("postProcessor Failed") // TODO: warning instead
				}
			}

		}

		rawReq, scheme, host, err = generator.next(data)

		if err != nil {
			return err // TODO: warning instead
		}

	}

	return nil
}

// TODO multitheading

// TODO finish

type FuzzGroupGenerator interface {
	next(interface{}) (FuzzArgGenerator, FuzzPostProcessor, interface{}, error)
}

type FuzzGroupPostProcessor interface {
	process(interface{}) error
}

func groupFuzz(groupGenerator FuzzGroupGenerator, groupPostProcessor FuzzGroupPostProcessor, groupData interface{}) error {
	generator, processor, data, err := groupGenerator.next(groupData)
	if err != nil {
		return err // TODO: warning instead
	}

	for generator != nil {

		// postProcessor can be nil ( is optional)
		//
		// if processor == nil {
		//		return errors.New("generator exists even tho processor is nil") // TODO: warning instead???
		// }
		//  data can be nil
		//
		//  if data == nil{
		// 	    return errors.New("generator and processor exists even tho data is nil")
		//  }

		err = fuzz(generator, processor, data)
		if err != nil {
			return err // TODO: warning instead
		}

		if groupPostProcessor != nil {
			groupPostProcessor.process(data)
		}

		generator, processor, data, err = groupGenerator.next(groupData)
		if err != nil {
			return err // TODO: warning instead
		}

	}

	return nil
}

func runTaskAsync(task interface {
	next(interface{}, interface{}) (bool, interface{}, error)
}, args interface{}, shared interface{}, threadNum *int, rate *int) error {

	if threadNum == nil {
		threadNum = new(int)
		*threadNum = 1
	}
	if rate == nil {
		rate = new(int)
		*rate = 1000000
	}
	ratemicros := 1000000 / *rate
	ratelimiter := time.Ticker{}
	ratelimiter = *time.NewTicker(time.Microsecond * time.Duration(ratemicros))
	threadlimiter := make(chan bool, *threadNum)
	//rateMutex := sync.Mutex{}
	var wg sync.WaitGroup
	var rateCounter *ring.Ring
	rateCounter = ring.New(*threadNum * 5)

	finished := false

	for !finished {
		threadlimiter <- true
		<-ratelimiter.C

		wg.Add(1)

		go func() {
			defer func() { <-threadlimiter }()
			defer wg.Done()
			// threadStart := time.Now()
			//j.runTask(nextInput, nextPosition, false)
			resp, i, err := task.next(args, shared)
			if err != nil {
				fmt.Printf("runTaskAsync err: %s", err)
			}
			if i != nil {
				// nothin
			}
			finished = resp
			// delay configuration: j.sleepIfNeeded()
			threadEnd := time.Now()
			//rateMutex.Lock()
			//defer rateMutex.Unlock()
			rateCounter = rateCounter.Next()
			rateCounter.Value = threadEnd.UnixMicro()
		}()
	}
	wg.Wait()
	return nil

}

type FuzzArgPacked struct {
	generator     FuzzArgGenerator
	postProcessor FuzzPostProcessor
}

type FuzzAsyncTask struct {
}

func (f FuzzAsyncTask) next(iArgsPacked interface{}, data interface{}) (bool, interface{}, error) {

	argsPacked, succ := iArgsPacked.(FuzzArgPacked)
	if succ != true {
		return false, nil, errors.New("FuzzArgPacked is not first argument of FuzzAsyncTask.next(interface{}, interface{}) method")
	}
	generator, postProcessor := argsPacked.generator, argsPacked.postProcessor

	rawReq, scheme, host, err := generator.next(data)

	// finish condition
	if rawReq == nil {
		// finish
		return true, nil, nil // end
	}
	// end

	if err != nil {
		return false, nil, err
	}

	if scheme == nil {
		return false, nil, errors.New("rawReq exists, even tho scheme doesn't exisit") // TODO: warning instead?
	}

	httpReq, err := makeHttpReq(*rawReq, scheme, host)
	if err != nil {
		return false, nil, err // TODO: warning instead
	}

	transport := http.Transport{DisableKeepAlives: false, DisableCompression: true} // disable all parameters that modifies request (can't edit "Proxy-Authorization" handling)

	client := http.Client{}

	client.Transport = &transport

	httpResp, err := client.Do(httpReq)

	if err != nil {
		return false, nil, err // TODO: warning instead
	}

	if postProcessor != nil {

		if httpResp == nil {
			return false, nil, errors.New("httpResp is null") // TODO: warning instead
		} else {
			err = postProcessor.process(*httpResp, data)
			if err != nil {
				return false, nil, errors.New("postProcessor Failed") // TODO: warning instead
			}
		}

	}

	return false, nil, nil
}

func fuzzAsync(generator FuzzArgGenerator, postProcessor FuzzPostProcessor, data interface{}, threadNum *int, rate *int) error {

	argPacked := FuzzArgPacked{generator: generator, postProcessor: postProcessor}

	err := runTaskAsync(FuzzAsyncTask{}, argPacked, data, threadNum, rate)

	if err != nil {
		return err
	}
	return nil
}
