# Sistema de inventario y abastecimiento (DDD) — Go

Proyecto de modelado de dominio para una cadena de supermercados.

## Ejecutar
    go run .        # demostración de escenarios
    go test ./...   # pruebas unitarias

## Estructura
- `dominio/`  modelo de dominio puro (sin dependencias de infraestructura)
- `memoria/`  repositorios en memoria (infraestructura)
- `main.go`   escenarios de demostración
