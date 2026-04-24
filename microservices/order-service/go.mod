module github.com/microservices/order-service

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/microservices/shared v0.0.0
)

replace github.com/microservices/shared => ../shared
