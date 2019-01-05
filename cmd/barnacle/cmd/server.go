package cmd

import (
	"net/http"

	"github.com/palestamp/barnacle/pkg/service"

	"github.com/spf13/cobra"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/apis"
	"github.com/palestamp/barnacle/pkg/backends"
	"github.com/palestamp/barnacle/pkg/backends/postgres"
	"github.com/palestamp/barnacle/pkg/metadata"
)

var (
	scServerAddr  string
	scPostgresURI string
)

const scPostgresURIDefault = "postgresql://postgres@localhost:5432/barnacle"

func ServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run Barnacle server.",
		RunE:  serveCmd,
	}

	cmd.Flags().StringVar(&scServerAddr, "addr", ":9878", "Address to listen on, ex: localhost:9878")
	cmd.Flags().StringVar(&scPostgresURI, "metadata-uri", scPostgresURIDefault, "Address of postgres metadata storage")
	return cmd
}

func serveCmd(cmd *cobra.Command, args []string) error {
	metadataStorage, err := metadata.NewPostgresStorage(scPostgresURI)
	if err != nil {
		return err
	}
	mds := metadata.WithLogging(metadataStorage)

	proxy := backends.NewRegistry()
	proxy.RegisterConnector(api.BackendType("postgres"), postgres.NewConnector())

	svc := service.New(proxy, mds)

	server := &http.Server{
		Handler: apis.NewV1API(svc),
		Addr:    scServerAddr,
	}

	return server.ListenAndServe()
}
