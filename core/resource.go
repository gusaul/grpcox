package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// BasePath define path where proto file will persisted
const BasePath = "/tmp/grpcox/"

// Resource - hold 3 main function (List, Describe, and Invoke)
type Resource struct {
	clientConn *grpc.ClientConn
	descSource grpcurl.DescriptorSource
	refClient  *grpcreflect.Client
	protos     []Proto
	protosets  []Proto

	headers []string
	md      metadata.MD
}

//openDescriptor - use it to reflect server descriptor
func (r *Resource) openDescriptor() error {
	ctx := context.Background()
	refCtx := metadata.NewOutgoingContext(ctx, r.md)
	r.refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(r.clientConn))

	// if no protos available use server reflection
	if r.protos == nil && r.protosets == nil {
		r.descSource = grpcurl.DescriptorSourceFromServer(ctx, r.refClient)
		return nil
	}

	protoPath := filepath.Join(BasePath, r.clientConn.Target())

	var err error
	if len(r.protosets) > 0 {
		// make list of protos name to be used as descriptor
		protos := make([]string, 0, len(r.protosets))
		for _, proto := range r.protosets {
			protos = append(protos, filepath.Join(protoPath, proto.Name))
		}

		r.descSource, err = grpcurl.DescriptorSourceFromProtoSets(protos...)
	} else {
		// make list of protos name to be used as descriptor
		protos := make([]string, 0, len(r.protos))
		for _, proto := range r.protos {
			protos = append(protos, proto.Name)
		}

		r.descSource, err = grpcurl.DescriptorSourceFromProtoFiles([]string{protoPath}, protos...)
	}
	return err
}

//closeDescriptor - please ensure to always close after open in the same flow
func (r *Resource) closeDescriptor() {
	done := make(chan int)
	go func() {
		if r.refClient != nil {
			r.refClient.Reset()
		}
		done <- 1
	}()

	select {
	case <-done:
		return
	case <-time.After(3 * time.Second):
		log.Printf("Reflection %s failed to close\n", r.clientConn.Target())
		return
	}
}

// List - To list all services exposed by a server
// symbol can be "" to list all available services
// symbol also can be service name to list all available method
func (r *Resource) List(symbol string) ([]string, error) {
	err := r.openDescriptor()
	if err != nil {
		return nil, err
	}
	defer r.closeDescriptor()

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
	err := r.openDescriptor()
	if err != nil {
		return "", "", err
	}
	defer r.closeDescriptor()

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
func (r *Resource) Invoke(ctx context.Context, metadata []string, symbol string, in io.Reader) (string, time.Duration, error) {
	err := r.openDescriptor()
	if err != nil {
		return "", 0, err
	}
	defer r.closeDescriptor()

	var resultBuffer bytes.Buffer

	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.Format("json"), r.descSource, false, true, in)
	if err != nil {
		return "", 0, err
	}
	h := grpcurl.NewDefaultEventHandler(&resultBuffer, r.descSource, formatter, false)

	var headers = r.headers
	if len(metadata) != 0 {
		headers = metadata
	}

	start := time.Now()
	err = grpcurl.InvokeRPC(ctx, r.descSource, r.clientConn, symbol, headers, h, rf.Next)
	end := time.Now().Sub(start) / time.Millisecond
	if err != nil {
		return "", end, err
	}

	if h.Status.Code() != codes.OK {
		return "", end, fmt.Errorf(h.Status.Message())
	}

	return resultBuffer.String(), end, nil
}

// Close - to close all resources that was opened before
func (r *Resource) Close() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if r.clientConn != nil {
			r.clientConn.Close()
			r.clientConn = nil
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := os.RemoveAll(BasePath)
		if err != nil {
			log.Printf("error removing proto dir from tmp: %s", err.Error())
		}
	}()

	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return
	case <-time.After(3 * time.Second):
		log.Printf("Connection %s failed to close\n", r.clientConn.Target())
		return
	}
}

func (r *Resource) isValid() bool {
	return r.refClient != nil && r.clientConn != nil
}

func (r *Resource) exit(code int) {
	// to force reset before os exit
	r.Close()
	os.Exit(code)
}

// AddProtos to resource properties and harddisk
// added protos will be persisted in `basepath + connection target`
// i.e. connection target == 127.0.0.1:8888
// proto files will be persisted in /tmp/grpcox/127.0.0.1:8888
// if the directory is already there, remove it first
func (r *Resource) AddProtos(protos []Proto) error {
	protoPath := filepath.Join(BasePath, r.clientConn.Target())
	err := os.MkdirAll(protoPath, 0777)
	if os.IsExist(err) {
		os.RemoveAll(protoPath)
		err = os.MkdirAll(protoPath, 0777)
	} else if err != nil {
		return err
	}

	var protoSlice, protosetSlice []Proto
	for _, proto := range protos {
		var err error
		if strings.HasSuffix(proto.Name, ".protoset") {
			protosetSlice = append(protosetSlice, proto)
			err = ioutil.WriteFile(filepath.Join(protoPath, "/", proto.Name),
				proto.Content,
				0777)
		} else {
			protoSlice = append(protoSlice, proto)
			err = ioutil.WriteFile(filepath.Join(protoPath, "/", proto.Name),
				prepareImport(proto.Content),
				0777)
		}
		if err != nil {
			return err
		}
	}

	r.protos = protoSlice
	r.protosets = protosetSlice
	return nil
}

// prepareImport transforming proto import into local path
// with exception to google proto import as it won't cause any problem
func prepareImport(proto []byte) []byte {
	const pattern = `import ".+`
	result := string(proto)

	re := regexp.MustCompile(pattern)
	matchs := re.FindAllString(result, -1)
	for _, match := range matchs {
		if strings.Contains(match, "\"google/") {
			continue
		}
		name := strings.Split(match, "/")
		if len(name) < 2 {
			continue
		}
		importString := `import "` + name[len(name)-1]
		result = strings.Replace(result, match, importString, -1)
	}

	return []byte(result)
}
