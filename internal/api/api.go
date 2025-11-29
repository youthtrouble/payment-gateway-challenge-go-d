package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

type Api struct {
	router         *chi.Mux
	paymentService *service.PaymentService
}

func New() *Api {
	return NewWithBankURL("http://localhost:8081")
}

func NewWithBankURL(bankURL string) *Api {
	// Initialize dependencies from bottom up
	repo := repository.NewPaymentsRepository()
	bankClient := client.NewHTTPBankClient(bankURL)
	paymentService := service.NewPaymentService(bankClient, repo)

	a := &Api{
		paymentService: paymentService,
	}
	a.setupRouter()

	return a
}

func (a *Api) Run(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:        addr,
		Handler:     a.router,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		fmt.Printf("shutting down HTTP server\n")
		return httpServer.Shutdown(ctx)
	})

	g.Go(func() error {
		fmt.Printf("starting HTTP server on %s\n", addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *Api) setupRouter() {
	a.router = chi.NewRouter()
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer) // Recover from panics

	a.router.Get("/ping", a.PingHandler())
	a.router.Get("/swagger/*", a.SwaggerHandler())

	a.router.Post("/api/payments", a.PostPaymentHandler())
	a.router.Get("/api/payments/{id}", a.GetPaymentHandler())
}

func (a *Api) Router() *chi.Mux {
	return a.router
}
