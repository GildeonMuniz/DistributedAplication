module github.com/microservices/user-service

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/microservices/shared v0.0.0
	golang.org/x/crypto v0.27.0
)

replace github.com/microservices/shared => ../shared
