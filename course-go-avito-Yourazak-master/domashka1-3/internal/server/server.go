package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func Start(ctx context.Context, handler http.Handler, port string) {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		fmt.Println("Server started on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Listen error: %v\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Println()
	fmt.Println("Shutting down service-courier")
	fmt.Println()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
}
