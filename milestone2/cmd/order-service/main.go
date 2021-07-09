package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/confluentinc/confluent-kafka-go/kafka"

)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})
	r.Get("/status", func(writer http.ResponseWriter, request *http.Request) {
		render.Render(writer, request,newStatusResponse())
	})
	r.Route("/orders", func(r chi.Router) {
		r.Use(ProducerCtx(p))

		r.Post("/", CreateOrder)
	})

	http.ListenAndServe("localhost:3000", r)
}

type ProductRequest struct {
	ProductCode string `json:"productCode"`
	Quantity int `json:"quantity"`
}

type Address struct {
	Line1 string `json:"line1"`
	City string `json:"city"`
	State string `json:"state"`
	PostalCode string `json:"postalCode"`
}
type CustomerRequest struct {
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
	ShippingAddress Address `json:"shippingAddress"`
}
type OrderRequest struct {
	Products []ProductRequest `json:"products"`
	Customer CustomerRequest `json:"customer"`
}
func (o *OrderRequest) Bind(r *http.Request) error {
	//if o.Customer == nil {
	//	return errors.New("missing required customer fields")
	//}
	return nil
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	data := &OrderRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	topic := "order-received"
	fmt.Printf("%+v\n", *data)
	json, _ := json.Marshal(data)

	p := r.Context().Value("producer").(*kafka.Producer);
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          json,
	}, nil)
	render.Render(w, r, newStatusResponse())

}

func ProducerCtx(producer *kafka.Producer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "producer", producer);
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

type StatusResponse struct {
	Status string
}

func newStatusResponse() *StatusResponse {
	resp := &StatusResponse{Status: "healthy!"}
	return resp
}
func (nsr *StatusResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}


type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
