package main

import (
	"context"
	"fmt"
	"log"
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
	"google.golang.org/grpc/peer"
)

func main() {
	var printCerts bool
	var myName string
	var serverEndpoint string
	var serverID string
	var delayBetweenWritesInSeconds int

	app := &cli.App{
		Name: "client",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "printcerts",
				EnvVars:     []string{"PRINT_CERTS"},
				Destination: &printCerts,
			},
			&cli.StringFlag{
				Name:        "name",
				EnvVars:     []string{"NAME"},
				Value:       "client",
				Destination: &myName,
			},
			&cli.StringFlag{
				Name:        "serverendpoint",
				EnvVars:     []string{"SERVER_ENDPOINT"},
				Value:       "server:443",
				Destination: &serverEndpoint,
			},
			&cli.StringFlag{
				Name:        "serverid",
				EnvVars:     []string{"SERVER_ID"},
				Value:       "spiffe://arunsworld.com/ns/ennovation/sa/default/name/server",
				Destination: &serverID,
			},
			&cli.IntFlag{
				Name:        "delay",
				EnvVars:     []string{"DELAY_BETWEEN_WRITES"},
				Value:       10, // 10 seconds
				Destination: &delayBetweenWritesInSeconds,
			},
		},
		Action: func(*cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			return doMain(ctx, myName, printCerts, serverEndpoint, serverID, time.Duration(delayBetweenWritesInSeconds)*time.Second)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func doMain(ctx context.Context, myName string, printCerts bool, serverEndpoint, _serverID string, delay time.Duration) error {
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

	serverID := spiffeid.RequireFromString(_serverID)
	conn, err := grpc.DialContext(ctx, serverEndpoint, grpc.WithTransportCredentials(
		grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeID(serverID)),
	))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)

	for {
		p := new(peer.Peer)
		reply, err := client.SayHello(ctx, &helloworld.HelloRequest{Name: myName}, grpc.Peer(p))
		if err != nil {
			return fmt.Errorf("error connecting to server %v", err)
		}
		serverID := "SOME SERVER"
		if peerID, ok := grpccredentials.PeerIDFromPeer(p); ok {
			serverID = peerID.String()
		}
		log.Printf("Reply from %s: %s", serverID, reply.Message)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(delay):
		}
	}
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
