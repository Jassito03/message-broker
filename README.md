# Universidad Nacional de Costa Rica
# Segunta tarea programada de Sistemas Operativos
**Curso:** EIF-212: Sistemas Operativos

**Profesor:** Jose Pablo Calvo Suarez

**Autores**
1. Brandon Castillo Badilla
2. Joseph Piñar Baltodano

## Objetivo:

Crear un Message broker utilizando un servidor GO con gRPC y un cliente go.

## Comandos importantes

### Compilar proto
```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/forum.proto
```
### Iniciar la app
```
go run server/main.go
go run client/main.go
```
