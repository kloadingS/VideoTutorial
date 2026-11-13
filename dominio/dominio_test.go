package dominio_test

import (
	"errors"
	"testing"
	"time"

	"supermercado/dominio"
	"supermercado/memoria"
)

const (
	arroz  = dominio.SKU("ARR-1KG")
	leche  = dominio.SKU("LEC-1L")
	aceite = dominio.SKU("ACE-1L")
	cafe   = dominio.SKU("CAF-500")
)

func relojSimulado() func() time.Time {
	t := time.Date(2026, 11, 2, 8, 0, 0, 0, time.UTC)
	return func() time.Time { t = t.Add(time.Hour); return t }
}

type mundo struct {
	productos   *memoria.Productos
	proveedores *memoria.Proveedores
	solicitudes *memoria.Solicitudes
	servicio    *dominio.ServicioAbastecimiento
	alajuela    *dominio.Inventario
}

// must exige que la construcción haya sido válida (solo para armar datos de prueba).
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func montar(t *testing.T) *mundo {
	t.Helper()
	m := &mundo{productos: memoria.NuevosProductos(), proveedores: memoria.NuevosProveedores(), solicitudes: memoria.NuevasSolicitudes()}
	pres := func(n string, u int) dominio.Presentacion { return must(dominio.NuevaPresentacion(n, u)) }
	unidad := pres("Unidad", 1)
	for _, d := range []struct {
		sku  dominio.SKU
		n    string
		caja dominio.Presentacion
	}{{arroz, "Arroz 1kg", pres("Caja x20", 20)}, {leche, "Leche 1L", pres("Caja x12", 12)}, {aceite, "Aceite 1L", pres("Caja x6", 6)}, {cafe, "Café 500g", pres("Caja x10", 10)}} {
		m.productos.Guardar(must(dominio.NuevoProducto(d.sku, d.n, "Abarrotes", unidad, d.caja)))
	}
	cond := func(p *dominio.Proveedor, sku dominio.SKU, pr string, precio int64, min, max int) {
		p.AgregarCondicion(must(dominio.NuevaCondicion(sku, pr, must(dominio.Colones(precio)), min, max)))
	}
	p1 := must(dominio.NuevoProveedor("P1", "Distribuidora Central"))
	cond(p1, arroz, "Caja x20", 16000, 2, 50) // 800 por unidad
	cond(p1, leche, "Caja x12", 9600, 1, 20)  // 800 por unidad
	cond(p1, aceite, "Caja x6", 15000, 1, 10)
	p2 := must(dominio.NuevoProveedor("P2", "Lácteos del Norte"))
	cond(p2, arroz, "Caja x20", 17000, 1, 100) // 850 por unidad
	cond(p2, leche, "Caja x12", 8400, 3, 10)   // 700 por unidad: el más barato
	m.proveedores.Guardar(p1)
	m.proveedores.Guardar(p2)
	m.servicio = dominio.NuevoServicioAbastecimiento(m.productos, m.proveedores, m.solicitudes)

	suc := must(dominio.NuevaSucursal("ALA", "Sucursal Alajuela"))
	m.alajuela = suc.NuevoInventario(relojSimulado())
	for _, c := range []struct {
		sku             dominio.SKU
		min, obj, stock int
	}{{arroz, 20, 100, 15}, {leche, 12, 60, 10}, {aceite, 6, 24, 6}, {cafe, 5, 20, 2}} {
		if err := m.alajuela.RegistrarProducto(c.sku, c.min, c.obj); err != nil {
			t.Fatal(err)
		}
		if err := m.alajuela.RegistrarEntrada(c.sku, c.stock, "Inventario inicial"); err != nil {
			t.Fatal(err)
		}
	}
	return m
}

func existencias(t *testing.T, inv *dominio.Inventario, sku dominio.SKU) int {
	return must(inv.Existencias(sku)).Valor()
}

// ---------- Value Objects ----------

func TestValueObjectsRechazanValoresInvalidos(t *testing.T) {
	if _, err := dominio.NuevoSKU("  "); !errors.Is(err, dominio.ErrSKUInvalido) {
		t.Errorf("SKU vacío debía fallar: %v", err)
	}
	if _, err := dominio.NuevaCantidad(-1); !errors.Is(err, dominio.ErrCantidadNegativa) {
		t.Errorf("cantidad negativa debía fallar: %v", err)
	}
	if _, err := dominio.NuevaPresentacion("Caja", 0); !errors.Is(err, dominio.ErrPresentacionInvalida) {
		t.Errorf("presentación con 0 unidades debía fallar: %v", err)
	}
	if _, err := dominio.NuevoProducto("X", "Sin presentaciones", "Abarrotes"); !errors.Is(err, dominio.ErrProductoInvalido) {
		t.Errorf("producto sin presentaciones debía fallar: %v", err)
	}
	if s := must(dominio.Colones(1250000)).String(); s != "₡1.250.000" {
		t.Errorf("formato de dinero: %s", s)
	}
}

// ---------- Inventario ----------

func TestSalidaMayorQueExistenciasSeRechazaYNoCambiaNada(t *testing.T) {
	m := montar(t)
	err := m.alajuela.RegistrarSalida(arroz, 16, "Venta")
	if !errors.Is(err, dominio.ErrExistenciasInsuficientes) {
		t.Fatalf("esperaba existencias insuficientes: %v", err)
	}
	if got := existencias(t, m.alajuela, arroz); got != 15 {
		t.Errorf("las existencias cambiaron: %d", got)
	}
	if n := len(m.alajuela.HistorialDe(arroz)); n != 1 {
		t.Errorf("no debía registrarse movimiento, hay %d", n)
	}
}

func TestCantidadesNoPositivasYProductoNoRegistrado(t *testing.T) {
	m := montar(t)
	if err := m.alajuela.RegistrarEntrada(arroz, 0, "x"); !errors.Is(err, dominio.ErrCantidadNoPositiva) {
		t.Errorf("entrada de 0: %v", err)
	}
	if err := m.alajuela.RegistrarSalida("NOEXISTE", 1, "x"); !errors.Is(err, dominio.ErrProductoNoRegistrado) {
		t.Errorf("producto no registrado: %v", err)
	}
	if err := m.alajuela.RegistrarProducto(arroz, 1, 5); !errors.Is(err, dominio.ErrProductoDuplicado) {
		t.Errorf("duplicado: %v", err)
	}
	if err := m.alajuela.ConfigurarPolitica(arroz, 10, 10); !errors.Is(err, dominio.ErrPoliticaInvalida) {
		t.Errorf("política con objetivo <= mínimo: %v", err)
	}
}

func TestAjusteExigeMotivoYQuedaEnHistorial(t *testing.T) {
	m := montar(t)
	if err := m.alajuela.Ajustar(arroz, 12, " "); !errors.Is(err, dominio.ErrMotivoRequerido) {
		t.Fatalf("ajuste sin motivo: %v", err)
	}
	if err := m.alajuela.Ajustar(arroz, 12, "Conteo físico: faltan 3"); err != nil {
		t.Fatal(err)
	}
	h := m.alajuela.HistorialDe(arroz)
	if last := h[len(h)-1]; last.Tipo() != dominio.Ajuste || last.Delta() != -3 {
		t.Errorf("movimiento de ajuste incorrecto: %v %d", last.Tipo(), last.Delta())
	}
	if err := m.alajuela.VerificarConsistencia(); err != nil {
		t.Error(err)
	}
}

func TestHistorialReconstruyeExistenciasEnUnaFecha(t *testing.T) {
	m := montar(t)
	// Reloj simulado: cada movimiento avanza 1 hora desde 2026-11-02 08:00.
	// Movimientos iniciales: 4 entradas (09:00..12:00). La de arroz es la #1 (09:00).
	m.alajuela.RegistrarEntrada(arroz, 50, "Compra") // 13:00
	m.alajuela.RegistrarSalida(arroz, 30, "Ventas")  // 14:00
	corte := time.Date(2026, 11, 2, 13, 30, 0, 0, time.UTC)
	if got := m.alajuela.ExistenciasEn(arroz, corte); got != 65 {
		t.Errorf("a las 13:30 había 65 y se calculó %d", got)
	}
	if got := existencias(t, m.alajuela, arroz); got != 35 {
		t.Errorf("hoy hay 35 y se calculó %d", got)
	}
}

func TestMovimientosDevueltosSonCopia(t *testing.T) {
	m := montar(t)
	cp := m.alajuela.Movimientos()
	n := len(cp)
	_ = append(cp[:0], cp[n-1]) // intento de alterar la copia
	if len(m.alajuela.Movimientos()) != n {
		t.Error("el historial original se alteró")
	}
}

func TestSucursalesTienenInventariosIndependientes(t *testing.T) {
	m := montar(t)
	heredia := must(dominio.NuevaSucursal("HER", "Sucursal Heredia")).NuevoInventario(nil)
	heredia.RegistrarProducto(arroz, 5, 50)
	heredia.RegistrarEntrada(arroz, 40, "Inicial")
	m.alajuela.RegistrarSalida(arroz, 5, "Venta")
	if existencias(t, heredia, arroz) != 40 || existencias(t, m.alajuela, arroz) != 10 {
		t.Error("las sucursales no deben compartir existencias")
	}
}

// ---------- Proveedor ----------

func TestProveedorValidaCondiciones(t *testing.T) {
	m := montar(t)
	p2 := must(m.proveedores.Obtener("P2"))
	prodAceite := must(m.productos.Obtener(aceite))
	prodLeche := must(m.productos.Obtener(leche))
	if _, err := p2.CrearLinea(prodAceite, 1); !errors.Is(err, dominio.ErrProveedorNoOfrece) {
		t.Errorf("P2 no vende aceite: %v", err)
	}
	if _, err := p2.CrearLinea(prodLeche, 2); !errors.Is(err, dominio.ErrCantidadFueraDeRango) {
		t.Errorf("mínimo de P2 es 3 cajas: %v", err)
	}
	if _, err := p2.CrearLinea(prodLeche, 11); !errors.Is(err, dominio.ErrCantidadFueraDeRango) {
		t.Errorf("máximo de P2 es 10 cajas: %v", err)
	}
	if _, err := dominio.NuevaCondicion(leche, "Caja x12", must(dominio.Colones(0)), 1, 5); !errors.Is(err, dominio.ErrCondicionInvalida) {
		t.Errorf("precio 0: %v", err)
	}
}

// ---------- Servicio de abastecimiento ----------

func TestGenerarSolicitudesEligeProveedorMasBaratoYAgrupa(t *testing.T) {
	m := montar(t)
	res, err := m.servicio.GenerarSolicitudes(m.alajuela)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Solicitudes) != 2 {
		t.Fatalf("esperaba 2 solicitudes (P1 y P2), hay %d", len(res.Solicitudes))
	}
	p1, p2 := res.Solicitudes[0], res.Solicitudes[1]
	if p1.ProveedorID() != "P1" || len(p1.Lineas()) != 2 { // arroz + aceite
		t.Errorf("P1 debía llevar arroz y aceite: %s %d líneas", p1.ProveedorID(), len(p1.Lineas()))
	}
	if p2.ProveedorID() != "P2" || p2.Lineas()[0].SKU() != leche {
		t.Errorf("la leche debía ir con P2 (más barato)")
	}
	cant := map[dominio.SKU]int{}
	for _, s := range res.Solicitudes {
		for _, l := range s.Lineas() {
			cant[l.SKU()] = l.Cantidad()
		}
	}
	// arroz: falta 85 -> ceil(85/20)=5 cajas; leche: falta 50 -> ceil(50/12)=5; aceite: falta 18 -> 3
	if cant[arroz] != 5 || cant[leche] != 5 || cant[aceite] != 3 {
		t.Errorf("cantidades inesperadas: %v", cant)
	}
	if len(res.SinProveedor) != 1 || res.SinProveedor[0] != cafe {
		t.Errorf("el café no tiene proveedor: %v", res.SinProveedor)
	}
	if got := p1.Total().Colones(); got != 5*16000+3*15000 {
		t.Errorf("total P1: %d", got)
	}
}

func TestGenerarSolicitudesAjustaAlMinimoDelProveedor(t *testing.T) {
	m := montar(t)
	heredia := must(dominio.NuevaSucursal("HER", "Heredia")).NuevoInventario(nil)
	heredia.RegistrarProducto(leche, 12, 30)
	heredia.RegistrarEntrada(leche, 11, "Inicial") // falta 19 -> 2 cajas, pero P2 exige 3
	res, err := m.servicio.GenerarSolicitudes(heredia)
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Solicitudes[0].Lineas()[0].Cantidad(); got != 3 {
		t.Errorf("debía subir al mínimo de 3 cajas, fue %d", got)
	}
}

func TestNoDuplicaSolicitudesPendientes(t *testing.T) {
	m := montar(t)
	m.servicio.GenerarSolicitudes(m.alajuela)
	res, err := m.servicio.GenerarSolicitudes(m.alajuela)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Solicitudes) != 0 || len(res.Omitidos) != 3 {
		t.Errorf("no debía generar nuevas solicitudes: %d nuevas, %d omitidos", len(res.Solicitudes), len(res.Omitidos))
	}
}

func TestCicloDeVidaYRecepcionActualizaInventario(t *testing.T) {
	m := montar(t)
	res, _ := m.servicio.GenerarSolicitudes(m.alajuela)
	sol := res.Solicitudes[1] // P2: 5 cajas de leche = 60 unidades

	if _, err := sol.Recibir(); !errors.Is(err, dominio.ErrTransicionInvalida) {
		t.Errorf("no se puede recibir un borrador: %v", err)
	}
	if err := sol.Enviar(); err != nil {
		t.Fatal(err)
	}
	if err := sol.Enviar(); !errors.Is(err, dominio.ErrTransicionInvalida) {
		t.Errorf("no se envía dos veces: %v", err)
	}
	if err := m.servicio.RecibirSolicitud(sol, m.alajuela); err != nil {
		t.Fatal(err)
	}
	if got := existencias(t, m.alajuela, leche); got != 10+60 {
		t.Errorf("leche debía quedar en 70: %d", got)
	}
	if err := sol.Cancelar(); !errors.Is(err, dominio.ErrTransicionInvalida) {
		t.Errorf("no se cancela una solicitud recibida: %v", err)
	}
	if err := m.alajuela.VerificarConsistencia(); err != nil {
		t.Error(err)
	}
}

func TestSolicitudSinLineasYLineaAjena(t *testing.T) {
	m := montar(t)
	sol := must(dominio.NuevaSolicitud("S-1", "ALA", "P1"))
	if err := sol.Enviar(); !errors.Is(err, dominio.ErrSolicitudSinLineas) {
		t.Errorf("enviar vacía: %v", err)
	}
	p2 := must(m.proveedores.Obtener("P2"))
	linea := must(p2.CrearLinea(must(m.productos.Obtener(leche)), 3))
	if err := sol.AgregarLinea(linea); !errors.Is(err, dominio.ErrLineaAjena) {
		t.Errorf("línea de otro proveedor: %v", err)
	}
}
