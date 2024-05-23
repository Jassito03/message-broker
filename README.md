# Segunta tarea programada de Sistemas Operativos

Message broker utilizando un servidor GO con gRPC y un cliente ...

## Comandos importantes

### Compilar proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/wishlist.proto
