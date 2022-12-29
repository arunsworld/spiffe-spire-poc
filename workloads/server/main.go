package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	const port = 50051
	var printCerts bool
	var trustDomain string

	app := &cli.App{
		Name: "client",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "printcerts",
				EnvVars:     []string{"PRINT_CERTS"},
				Value:       true,
				Destination: &printCerts,
			},
			&cli.StringFlag{
				Name:        "trustdomain",
				EnvVars:     []string{"TRUST_DOMAIN"},
				Value:       "arunsworld.com",
				Destination: &trustDomain,
			},
		},
		Action: func(*cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			return doMain(ctx, port, trustDomain, printCerts)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func doMain(ctx context.Context, port int, _trustDomain string, printCerts bool) error {
	log.Println("opening SPIFFE Workload API X.509 source...")

	certctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	source, err := workloadapi.NewX509Source(certctx)
	if err != nil {
		return err
	}
	defer source.Close()

	if printCerts {
		if err := printCertsForDebugging(source); err != nil {
			return err
		}
	}

	trustDomain := spiffeid.RequireTrustDomainFromString(_trustDomain)
	creds := grpccredentials.MTLSServerCredentials(source, source, tlsconfig.AuthorizeMemberOf(trustDomain))

	s := grpc.NewServer(grpc.Creds(creds))
	helloworld.RegisterGreeterServer(s, server{})
	defer s.Stop()

	endpoint := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}
	log.Println("starting gRPC server on: ", endpoint)
	go func() {
		<-ctx.Done()
		stopped := make(chan struct{})
		go func() {
			forceStopCtx, forceCancel := context.WithTimeout(context.Background(), time.Second*5)
			defer forceCancel()
			select {
			case <-stopped:
				return
			case <-forceStopCtx.Done():
				log.Println("forcing stop after timeout")
				s.Stop()
			}
		}()
		s.GracefulStop()
		close(stopped)
	}()
	if err := s.Serve(lis); err != nil {
		return err
	}
	log.Println("finished serving...")

	return nil
}

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s server) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	clientID := "SOME CLIENT"
	if peerID, ok := grpccredentials.PeerIDFromContext(ctx); ok {
		clientID = peerID.String()
	}
	log.Printf("received hello request from: %s: for: %s", clientID, req.Name)
	resp := fmt.Sprintf("On behalf of %s, Hello, %s", clientID, req.Name)
	return &helloworld.HelloReply{Message: resp}, nil
}

func printCertsForDebugging(source *workloadapi.X509Source) error {
	svid, err := source.GetX509SVID()
	if err != nil {
		return err
	}
	a, b, err := svid.Marshal()
	if err != nil {
		return err
	}
	fmt.Println(string(a))
	fmt.Println(string(b))
	return nil
}
