package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/urfave/cli/v2"
)

//go:embed all:embed/*
var webContent embed.FS

func main() {
	const port = 50051
	var printCerts bool

	app := &cli.App{
		Name: "client",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "printcerts",
				EnvVars:     []string{"PRINT_CERTS"},
				Value:       true,
				Destination: &printCerts,
			},
		},
		Action: func(*cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			return doMain(ctx, port, printCerts)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func doMain(ctx context.Context, port int, printCerts bool) error {
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

	indexHTML, err := readEmbedContent("index.html")
	if err != nil {
		return err
	}
	svid, err := source.GetX509SVID()
	if err != nil {
		return err
	}
	idMsg := fmt.Sprintf("I am: %s", svid.ID.String())
	indexHTML = bytes.ReplaceAll(indexHTML, []byte("I'm a secure webapp"), []byte(idMsg))
	indexCSS, err := readEmbedContent("index.css")
	if err != nil {
		return err
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(indexHTML)
	})
	http.HandleFunc("/index.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		w.Write(indexCSS)
	})

	tlsConfig := tlsconfig.TLSServerConfig(source)

	endpoint := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:      endpoint,
		TLSConfig: tlsConfig,
	}

	serverErr := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-serverErr:
			return
		}
		log.Println("graceful shutdown initiated...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("error during graceful shutdown: %v", err)
		}
	}()

	log.Printf("https server starting at endpoint: %v", endpoint)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		close(serverErr)
		return err
	}
	log.Println("https server shutdown...")

	return nil

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

func readEmbedContent(fname string) ([]byte, error) {
	return webContent.ReadFile(fmt.Sprintf("embed/%s", fname))
}
