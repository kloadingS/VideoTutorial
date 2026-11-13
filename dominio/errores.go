package dominio

import "errors"

// Errores de dominio. Se envuelven con fmt.Errorf("%w ...") para dar contexto
// y se pueden comparar con errors.Is.
var (
	ErrSKUInvalido              = errors.New("SKU inválido")
	ErrCantidadNegativa         = errors.New("la cantidad no puede ser negativa")
	ErrCantidadNoPositiva       = errors.New("la cantidad debe ser mayor que cero")
	ErrDineroInvalido           = errors.New("monto de dinero inválido")
	ErrPresentacionInvalida     = errors.New("presentación inválida")
	ErrPresentacionDesconocida  = errors.New("el producto no tiene esa presentación")
	ErrProductoInvalido         = errors.New("producto inválido")
	ErrProductoDuplicado        = errors.New("el producto ya está registrado")
	ErrProductoNoRegistrado     = errors.New("el producto no está registrado en el inventario")
	ErrExistenciasInsuficientes = errors.New("existencias insuficientes")
	ErrPoliticaInvalida         = errors.New("política de reposición inválida")
	ErrMotivoRequerido          = errors.New("el ajuste requiere un motivo")
	ErrInconsistencia           = errors.New("el historial no coincide con las existencias")
	ErrCondicionInvalida        = errors.New("condición de suministro inválida")
	ErrProveedorInvalido        = errors.New("proveedor inválido")
	ErrProveedorNoOfrece        = errors.New("el proveedor no ofrece el producto")
	ErrCantidadFueraDeRango     = errors.New("cantidad fuera de las condiciones del proveedor")
	ErrSolicitudInvalida        = errors.New("solicitud inválida")
	ErrLineaDuplicada           = errors.New("el producto ya está en la solicitud")
	ErrLineaAjena               = errors.New("la línea pertenece a otro proveedor")
	ErrSolicitudSinLineas       = errors.New("la solicitud no tiene líneas")
	ErrTransicionInvalida       = errors.New("transición de estado no permitida")
	ErrSucursalInvalida         = errors.New("sucursal inválida")
	ErrNoEncontrado             = errors.New("no encontrado")
)
