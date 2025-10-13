package app

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	handlerDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/download_cv"
	handlerGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_content"
	handlerGetCvLink "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_cv_link"
	processDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/download_cv"
	processGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_content"
	processGetCvLink "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_cv_link"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/registry"
	serviceGRPC "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/content_service/grpc"
	serviceHTTP "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/content_service/http"
	serviceRabbitMQ "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/content_service/rabbitmq"
)

type App struct {
	httpServer *http.Server
	grpcConn   *grpc.ClientConn
	rabbit     *serviceRabbitMQ.Client
}

func Build(cfg *registry.Config) (*App, error) {
	maxRetries := 20
	retryDelay := 2 * time.Second
	var err error

	var grpcConn *grpc.ClientConn
	dialCtx, cancel := context.WithTimeout(context.Background(), time.Duration(maxRetries)*retryDelay)
	defer cancel()
	grpcConn, err = grpc.DialContext(dialCtx, cfg.Services.Content.GRPC.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	log.Println("INFO: successfully connected to gRPC server")
	contentGRPCClient := serviceGRPC.NewClient(grpcConn)

	var rabbitClient *serviceRabbitMQ.Client
	for i := 0; i < maxRetries; i++ {
		if rabbitClient, err = serviceRabbitMQ.NewClient(cfg.RabbitMQ.URL, cfg.RabbitMQ.Topology.Exchange); err == nil {
			log.Println("INFO: successfully connected to RabbitMQ")
			break
		}
		log.Printf("INFO: could not connect to RabbitMQ, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	if err != nil {
		return nil, err
	}

	downloadCvHTTPClient := serviceHTTP.NewClient(cfg.Services.Content.HTTP.Addr)

	getContentProcess := processGetContent.NewProcess(contentGRPCClient)
	getCvLinkProcess := processGetCvLink.NewProcess(rabbitClient, cfg.RabbitMQ.Topology.CVRequestRoutingKey)
	downloadCvProcess := processDownloadCv.NewProcess(downloadCvHTTPClient)

	getContentHandler := handlerGetContent.NewHandler(getContentProcess)
	getCvLinkHandler := handlerGetCvLink.NewHandler(getCvLinkProcess)
	downloadCvHandler := handlerDownloadCv.NewHandler(downloadCvProcess)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/content", getContentHandler.Handle)
	mux.HandleFunc("/api/v1/cv-request", getCvLinkHandler.Handle)
	mux.HandleFunc("/download/cv", downloadCvHandler.Handle)

	httpServer := &http.Server{
		Addr: ":" + cfg.Server.HTTPPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			mux.ServeHTTP(w, r)
		}),
	}

	return &App{
		httpServer: httpServer,
		grpcConn:   grpcConn,
		rabbit:     rabbitClient,
	}, nil
}

func (a *App) RunHTTP() error {
	log.Printf("INFO: HTTP server listening on %s", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) {
	log.Println("INFO: shutting down servers...")
	_ = a.httpServer.Shutdown(ctx)
	_ = a.grpcConn.Close()
	_ = a.rabbit.Close()
}
