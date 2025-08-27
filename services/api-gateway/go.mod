module github.com/3nd3r1/kubin/api-gateway

go 1.22.3

require (
	github.com/3nd3r1/kubin/shared v0.0.0
	github.com/go-chi/chi/v5 v5.2.2
	github.com/kelseyhightower/envconfig v1.4.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace github.com/3nd3r1/kubin/shared => ../shared
