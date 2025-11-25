module github.com/your-org/ecommerce-platform/services/payment

go 1.23.0

require (
	github.com/google/uuid v1.6.0
	github.com/your-org/ecommerce-platform/shared v0.0.0
	go.mongodb.org/mongo-driver v1.17.6
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace github.com/your-org/ecommerce-platform/shared => ../../shared
