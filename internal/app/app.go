package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	handlerDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/download_cv"
	handlerGetCaptcha "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_captcha"
	handlerGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_content"
	handlerGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_cv_token"
	handlerGetPow "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/get_pow"
	handlerVerifyCaptcha "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/handler/verify_captcha"
	processDownloadCv "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/download_cv"
	processGetCaptcha "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_captcha"
	processGetContent "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_content"
	processGetCvToken "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_cv_token"
	processGetPow "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/get_pow"
	processVerifyCaptcha "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/process/verify_captcha"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/registry"
	serviceCaptcha "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
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
	if cfg == nil || cfg.Infrastructure.Retry.MaxAttempts < 1 || cfg.Infrastructure.Retry.DelaySeconds < 0 || cfg.Server.HTTPPort == "" || cfg.Services.Content.GRPC.Addr == "" || cfg.Services.Content.HTTP.Addr == "" || cfg.Services.Captcha.HTTP.Addr == "" || cfg.RabbitMQ.URL == "" || cfg.RabbitMQ.Topology.Exchange == "" || cfg.RabbitMQ.Topology.CVRequestRoutingKey == "" {
		return nil, fmt.Errorf("invalid gateway configuration")
	}
	maxRetries := cfg.Infrastructure.Retry.MaxAttempts
	retryDelay := cfg.Infrastructure.Retry.DelaySeconds
	var err error

	var grpcConn *grpc.ClientConn
	grpcConn, err = grpc.NewClient(cfg.Services.Content.GRPC.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Println("INFO: content gRPC client initialized")
	cleanupGRPC := true
	defer func() {
		if cleanupGRPC {
			_ = grpcConn.Close()
		}
	}()
	contentGRPCClient := serviceGRPC.NewClient(grpcConn)

	var rabbitClient *serviceRabbitMQ.Client
	for i := range maxRetries {
		if rabbitClient, err = serviceRabbitMQ.NewClient(cfg.RabbitMQ.URL, cfg.RabbitMQ.Topology.Exchange); err == nil {
			log.Println("INFO: successfully connected to RabbitMQ")
			break
		}
		if i+1 < maxRetries {
			log.Printf("INFO: could not connect to RabbitMQ, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return nil, err
	}
	cleanupGRPC = false

	downloadCvHTTPClient := serviceHTTP.NewClient(cfg.Services.Content.HTTP.Addr)
	captchaHTTPClient := serviceCaptcha.NewClient(&http.Client{Timeout: 10 * time.Second}, cfg.Services.Captcha.HTTP.Addr)

	getContentProcess := processGetContent.NewProcess(contentGRPCClient)
	getCvLinkProcess := processGetCvToken.NewProcess(rabbitClient, cfg.RabbitMQ.Topology.CVRequestRoutingKey)
	downloadCvProcess := processDownloadCv.NewProcess(downloadCvHTTPClient)

	getPowProcess := processGetPow.NewProcess(captchaHTTPClient)
	getCaptchaProcess := processGetCaptcha.NewProcess(captchaHTTPClient)
	verifyCaptchaProcess := processVerifyCaptcha.NewProcess(captchaHTTPClient)

	getContentHandler := handlerGetContent.NewHandler(getContentProcess)
	getCvLinkHandler := handlerGetCvToken.NewHandler(getCvLinkProcess)
	downloadCvHandler := handlerDownloadCv.NewHandler(downloadCvProcess)
	getPowHandler := handlerGetPow.NewHandler(getPowProcess)
	getCaptchaHandler := handlerGetCaptcha.NewHandler(getCaptchaProcess)
	verifyCaptchaHandler := handlerVerifyCaptcha.NewHandler(verifyCaptchaProcess)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/content", getContentHandler.Handle)
	mux.HandleFunc("/api/v1/cv-request", getCvLinkHandler.Handle)
	mux.HandleFunc("/api/v1/download/cv", downloadCvHandler.Handle)
	mux.HandleFunc("/api/v1/pow", getPowHandler.Handle)
	mux.HandleFunc("/api/v1/captcha", getCaptchaHandler.Handle)
	mux.HandleFunc("/api/v1/captcha-verify", verifyCaptchaHandler.Handle)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Server.HTTPPort,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
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
