package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Resource - hold 3 main function (List, Describe, and Invoke)
type Resource struct {
	clientConn *grpc.ClientConn
	descSource grpcurl.DescriptorSource
	refClient  *grpcreflect.Client

	headers []string
}

// List - To list all services exposed by a server
// symbol can be "" to list all available services
// symbol also can be service name to list all available method
func (r *Resource) List(symbol string) ([]string, error) {

	var result []string
	if symbol == "" {
		svcs, err := grpcurl.ListServices(r.descSource)
		if err != nil {
			return result, err
		}
		if len(svcs) == 0 {
			return result, fmt.Errorf("No Services")
		}

		for _, svc := range svcs {
			result = append(result, fmt.Sprintf("%s\n", svc))
		}
	} else {
		methods, err := grpcurl.ListMethods(r.descSource, symbol)
		if err != nil {
			return result, err
		}
		if len(methods) == 0 {
			return result, fmt.Errorf("No Function") // probably unlikely
		}

		for _, m := range methods {
			result = append(result, fmt.Sprintf("%s\n", m))
		}
	}

	return result, nil
}

// Describe - The "describe" verb will print the type of any symbol that the server knows about
// or that is found in a given protoset file.
// It also prints a description of that symbol, in the form of snippets of proto source.
// It won't necessarily be the original source that defined the element, but it will be equivalent.
func (r *Resource) Describe(symbol string) (string, string, error) {

	var result, template string

	var symbols []string
	if symbol != "" {
		symbols = []string{symbol}
	} else {
		// if no symbol given, describe all exposed services
		svcs, err := r.descSource.ListServices()
		if err != nil {
			return "", "", err
		}
		if len(svcs) == 0 {
			log.Println("Server returned an empty list of exposed services")
		}
		symbols = svcs
	}
	for _, s := range symbols {
		if s[0] == '.' {
			s = s[1:]
		}

		dsc, err := r.descSource.FindSymbol(s)
		if err != nil {
			return "", "", err
		}

		txt, err := grpcurl.GetDescriptorText(dsc, r.descSource)
		if err != nil {
			return "", "", err
		}
		result = txt

		if dsc, ok := dsc.(*desc.MessageDescriptor); ok {
			// for messages, also show a template in JSON, to make it easier to
			// create a request to invoke an RPC
			tmpl := grpcurl.MakeTemplate(dsc)
			_, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.Format("json"), r.descSource, true, false, nil)
			if err != nil {
				return "", "", err
			}
			str, err := formatter(tmpl)
			if err != nil {
				return "", "", err
			}
			template = str
		}
	}

	return result, template, nil
}

// Invoke - invoking gRPC function
func (r *Resource) Invoke(ctx context.Context, symbol string, in io.Reader) (string, time.Duration, error) {
	// because of grpcurl directlu fmt.Printf on their invoke function
	// so we stub the Stdout using os.Pipe
	backUpStdout := os.Stdout
	defer func() {
		os.Stdout = backUpStdout
	}()

	f, w, err := os.Pipe()
	if err != nil {
		return "", 0, err
	}
	os.Stdout = w

	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.Format("json"), r.descSource, false, true, in)
	if err != nil {
		return "", 0, err
	}
	h := grpcurl.NewDefaultEventHandler(os.Stdout, r.descSource, formatter, false)

	start := time.Now()
	err = grpcurl.InvokeRPC(ctx, r.descSource, r.clientConn, symbol, r.headers, h, rf.Next)
	end := time.Now().Sub(start) / time.Millisecond
	if err != nil {
		return "", end, err
	}

	if h.Status.Code() != codes.OK {
		return "", end, fmt.Errorf(h.Status.Message())
	}

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, f)
		outC <- buf.String()
	}()

	w.Close()
	out := <-outC

	return out, end, nil
}

// Close - to close all resources that was opened before
func (r *Resource) Close() {
	if r.refClient != nil {
		r.refClient.Reset()
		r.refClient = nil
	}
	if r.clientConn != nil {
		r.clientConn.Close()
		r.clientConn = nil
	}
}

func (r *Resource) exit(code int) {
	// to force reset before os exit
	r.Close()
	os.Exit(code)
}
