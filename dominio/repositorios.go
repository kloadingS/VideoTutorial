package dominio

// Contratos de persistencia. Pertenecen al DOMINIO (qué se necesita guardar y
// buscar); la implementación vive en infraestructura (paquete memoria).
// Hay un repositorio por raíz de agregado.

type ProductoRepositorio interface {
	Guardar(*Producto) error
	Obtener(SKU) (*Producto, error)
}

type ProveedorRepositorio interface {
	Guardar(*Proveedor) error
	Obtener(id string) (*Proveedor, error)
	Todos() ([]*Proveedor, error)
}

type InventarioRepositorio interface {
	Guardar(*Inventario) error
	Obtener(sucursalID string) (*Inventario, error)
}

type SolicitudRepositorio interface {
	Guardar(*SolicitudAbastecimiento) error
	Obtener(id string) (*SolicitudAbastecimiento, error)
	PendientesDe(sucursalID string) ([]*SolicitudAbastecimiento, error)
}
