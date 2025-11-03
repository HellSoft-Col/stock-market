# 🥑 Guía de Uso - Plataforma de Trading de Aguacates

## 📋 Resumen Rápido

Esta guía te explica qué hace cada botón y dónde verás los resultados en la interfaz.

---

## 🚀 Pestaña: SIMULADOR

### Botón: "Create Market Order"
**¿Qué hace?**
- Crea una orden de compra o venta simulada con precio límite fijo
- Usa `debugMode: AUTO_ACCEPT` para ejecución inmediata
- NO afecta tu balance real (es simulación)

**📍 Dónde ver el resultado:**
- ✅ Notificación toast (esquina superior derecha)
- ✅ Pestaña "My Orders" en Trading
- ✅ Panel "Market Activity" en Dashboard
- ✅ Contador de órdenes activas en Dashboard

---

### Botón: "Auto MM" (Market Maker)
**¿Qué hace?**
- Inicia/detiene la creación automática de órdenes cada 3 segundos
- Crea órdenes aleatorias de compra/venta
- Productos: FOSFO, PITA, PALTA-OIL, GUACA, SEBO, H-GUACA
- Cantidades: 1-5 unidades
- Precios: aleatorios entre 10-30

**📍 Dónde ver el resultado:**
- ✅ Notificación toast confirmando inicio/parada
- ✅ Pestaña "Messages" en Trading (log de cada orden)
- ✅ Pestaña "My Orders" se llena automáticamente
- ✅ Panel "Market Activity" muestra la actividad

**💡 Consejo:** Úsalo para crear liquidez artificial en el mercado y probar tu estrategia de trading

---

### Botón: "Clear All"
**⚠️ Estado:** Función pendiente de implementación en el servidor

**¿Qué hace?**
- Cancelaría todas las órdenes simuladas
- Actualmente muestra mensaje informativo

**📍 Dónde ver el resultado:**
- ℹ️ Notificación toast informativa

---

## 📈 Pestaña: TRADING

### Botón: "Place Order"
**¿Qué hace?**
- Envía una orden REAL al mercado (MARKET o LIMIT)
- Afecta tu balance e inventario real
- Puede incluir mensaje personalizado

**📍 Dónde ver el resultado:**
1. **Inmediatamente:**
   - ✅ Notificación toast de confirmación
   - ✅ Pestaña "My Orders" (orden aparece como pendiente)
   
2. **Cuando se ejecuta:**
   - ✅ Pestaña "History" (nueva transacción)
   - ✅ Actualización de Balance en Dashboard
   - ✅ Actualización de Inventario en sidebar
   - ✅ Contador de "Fills" en Dashboard aumenta

---

### Botón: "Refresh Orders"
**¿Qué hace?**
- Solicita al servidor la lista actualizada de tus órdenes activas
- Sincroniza tu vista con el estado real del servidor

**📍 Dónde ver el resultado:**
- ✅ Lista "My Orders" se actualiza
- ✅ Contador de órdenes activas en Dashboard
- ✅ Notificación toast con cantidad de órdenes

---

### Botones: "Quick Buy/Sell FOSFO"
**¿Qué hace?**
- Atajo rápido para comprar/vender 5 unidades de FOSFO
- Usa precio de mercado (MARKET order)
- Equivale a crear una orden manual pero más rápido

**📍 Dónde ver el resultado:**
- Igual que "Place Order" (ver arriba)
- ⌨️ **Atajo de teclado:** Ctrl + B (compra)

---

### Botón: "Cancel All"
**⚠️ Estado:** Función pendiente de implementación en el servidor

**¿Qué hace?**
- Cancelaría todas tus órdenes activas
- Actualmente muestra mensaje informativo

**📍 Dónde ver el resultado:**
- ℹ️ Notificación toast informativa

---

## 🐛 Pestaña: DEBUG

### Botones: Error Injection (Balance, Product, Disconnect, Expire)
**¿Qué hace?**
- Simula diferentes tipos de errores del servidor
- Útil para probar cómo tu cliente maneja errores
- Tipos:
  - **Balance:** Simula saldo insuficiente
  - **Product:** Simula producto no autorizado
  - **Disconnect:** Simula desconexión del cliente
  - **Expire:** Simula oferta expirada

**📍 Dónde ver el resultado:**
- ✅ Notificación toast de error (roja)
- ✅ Pestaña "Messages" con detalle del error
- ✅ Posible mensaje de ERROR del servidor

---

### Sección: Production Test
**¿Qué hace?**
- Simula la producción de un producto
- Aumenta la cantidad de ese producto en tu inventario
- Útil para probar algoritmos de producción

**📍 Dónde ver el resultado:**
- ✅ Inventario en sidebar izquierdo (cantidad aumenta)
- ✅ Notificación toast confirmando producción
- ✅ Se ejecuta RESYNC automático después

---

### Botón: "Ping"
**¿Qué hace?**
- Envía mensaje PING al servidor
- Verifica conectividad y mide latencia

**📍 Dónde ver el resultado:**
- ✅ Notificación toast "Pong received"
- ✅ Pestaña "Messages" con timestamp
- ⏱️ **Tiempo esperado:** <100ms en servidor local

---

### Botón: "Resync"
**¿Qué hace?**
- Solicita resincronización completa de datos
- Actualiza inventario, balance, y estado general

**📍 Dónde ver el resultado:**
- ✅ Inventario actualizado en sidebar
- ✅ Balance actualizado en Dashboard
- ✅ Notificaciones toast para cada actualización

---

### Botón: "Orders"
**¿Qué hace?**
- Solicita lista completa de TODAS las órdenes activas en el mercado
- No solo las tuyas, sino de todos los equipos

**📍 Dónde ver el resultado:**
- ✅ Pestaña "Messages" con log
- ℹ️ Para ver solo tus órdenes, usa "Refresh Orders" en Trading

---

### Botón: "Sessions"
**¿Qué hace?**
- Muestra todos los clientes conectados al servidor
- Incluye: nombre de equipo, tipo de cliente, estado de autenticación

**📍 Dónde ver el resultado:**
- ✅ Panel "Market Activity" en Dashboard
- ✅ Tarjeta azul con lista de sesiones
- ✅ Notificación toast con cantidad de sesiones

---

### Botón: "My Performance"
**¿Qué hace?**
- Genera reporte detallado de tu rendimiento
- Incluye:
  - P&L (Profit & Loss)
  - ROI (Return on Investment %)
  - Total de trades
  - Volumen negociado
  - Ratio Buy/Sell
  - Ranking (si disponible)

**📍 Dónde ver el resultado:**
- ✅ Panel "Market Activity" en Dashboard
- ✅ Tarjeta verde con todas las estadísticas
- ✅ Notificación toast confirmando carga

---

### Botón: "Global Report"
**¿Qué hace?**
- Genera reporte global del mercado
- Incluye:
  - Duración del mercado
  - Total de trades globales
  - Volumen total
  - Top 3 traders con mejor ROI

**📍 Dónde ver el resultado:**
- ✅ Panel "Market Activity" en Dashboard
- ✅ Tarjeta púrpura con ranking global
- ✅ Notificación toast confirmando carga

---

## 📊 Otras Pestañas

### Tab: "Ticker"
**¿Qué muestra?**
- Precios en tiempo real de todos los productos
- Best Bid (mejor oferta de compra)
- Best Ask (mejor oferta de venta)
- Mid (precio medio)
- Spread (diferencia bid-ask)
- Volumen 24h

**🔄 Actualización:**
- Automática cuando el servidor envía mensajes TICKER
- No requiere botón de refresh

---

### Tab: "Order Book"
**¿Qué muestra?**
- Órdenes de compra (verdes) - top 10
- Órdenes de venta (rojas) - top 10
- Lista combinada de todas las órdenes del producto seleccionado

**🔧 Cómo usar:**
1. Selecciona un producto del dropdown
2. El libro de órdenes se actualiza automáticamente
3. Puedes ver qué equipos tienen órdenes activas

---

### Tab: "History"
**¿Qué muestra?**
- Historial completo de todas tus transacciones ejecutadas
- Para cada transacción:
  - Lado (BUY/SELL)
  - Cantidad
  - Producto
  - Precio unitario
  - Valor total
  - Contraparte (con quién tradeas)
  - Timestamp

**🔄 Actualización:**
- Automática cuando recibes mensajes FILL del servidor
- Las más recientes aparecen arriba

---

## ⌨️ Atajos de Teclado

| Atajo | Acción |
|-------|--------|
| `Ctrl + Alt + D` | Abrir pestaña Debug |
| `Ctrl + Alt + S` | Abrir pestaña Simulator |
| `Ctrl + Alt + T` | Abrir pestaña Trading |
| `Ctrl + B` | Compra rápida FOSFO (5 unidades) |
| `Ctrl + Shift + R` | Actualizar órdenes |

---

## 🎨 Sistema de Notificaciones Toast

Las notificaciones aparecen en la **esquina superior derecha** con colores según el tipo:

| Color | Tipo | Ejemplos |
|-------|------|----------|
| 🟢 Verde | Éxito | Orden creada, conexión exitosa, pong recibido |
| 🔴 Rojo | Error | Error de autenticación, conexión fallida |
| 🟡 Amarillo | Advertencia | Por favor autentícate primero, campo requerido |
| 🔵 Azul | Información | Cargando datos, procesando solicitud |

**Características:**
- Auto-desaparecen después de 3 segundos (configurable)
- Puedes cerrarlas manualmente con la X
- Se apilan verticalmente si hay varias
- Animación suave de entrada/salida

---

## 🔍 Panel "Market Activity"

Ubicación: **Dashboard principal, panel inferior**

Este panel muestra en tiempo real:
- Notificaciones de trades ejecutados
- Actualizaciones de estado de mercado
- Resultados de comandos (Sessions, Performance Reports)
- Últimas 10 actividades (para evitar sobrecarga)

**Color por tipo:**
- Verde: Trades exitosos
- Azul: Actualizaciones de mercado, sesiones
- Púrpura: Reportes globales
- Verde oscuro: Reportes personales

---

## ⚠️ Notas Importantes

### Diferencia: Simulador vs Trading Real

| Aspecto | Simulador | Trading Real |
|---------|-----------|--------------|
| Afecta balance | ❌ No | ✅ Sí |
| Tipo de orden | Solo LIMIT | MARKET y LIMIT |
| DebugMode | AUTO_ACCEPT | Normal |
| Propósito | Pruebas, liquidez | Trading competitivo |
| Mensajes | "Simulator order" | Personalizable |

### Limitaciones Conocidas
1. **Cancelación de órdenes no disponible** - El servidor aún no implementa CANCEL
2. **Auto MM usa precios aleatorios** - No basados en datos reales de mercado
3. **Order Book muestra top 10** - Por rendimiento
4. **Market Activity muestra últimas 10** - Para evitar sobrecarga

---

## 🆘 Troubleshooting

### "Not connected to server"
- **Causa:** WebSocket no conectado
- **Solución:** Haz clic en "Connect" en la pestaña Auth

### "Please authenticate first"
- **Causa:** No has hecho login
- **Solución:** Ingresa tu token y haz clic en "Login"

### "No active orders"
- **Causa:** No tienes órdenes pendientes
- **Solución:** Crea órdenes con "Place Order" o "Create Market Order"

### Las órdenes no aparecen en "My Orders"
- **Causa:** La lista no se ha actualizado
- **Solución:** Haz clic en "Refresh Orders"

### El Order Book está vacío
- **Causa:** No has seleccionado un producto
- **Solución:** Selecciona un producto del dropdown

---

## 📝 Flujo de Trabajo Típico

### Para Trading Normal:
1. Connect → Login (pestaña Auth)
2. Verificar Balance e Inventario (sidebar)
3. Ir a Trading tab
4. Crear orden con "Place Order"
5. Verificar en "My Orders"
6. Cuando se ejecute, ver en "History"

### Para Testing/Simulación:
1. Connect → Login (pestaña Auth)
2. Ir a Simulator tab
3. Activar "Auto MM" para generar actividad
4. Observar Market Activity
5. Ir a Ticker para ver precios actualizándose
6. Usar Debug tools para probar errores

### Para Monitoreo:
1. Connect → Login (pestaña Auth)
2. Dashboard para visión general
3. "Sessions" para ver quién está conectado
4. "My Performance" para tu rendimiento
5. "Global Report" para ranking
6. Ticker para precios en tiempo real
7. Order Book para ver profundidad de mercado

---

## 💡 Consejos Pro

1. **Usa Auto MM primero** para generar liquidez antes de tradear
2. **Monitorea el Ticker** antes de colocar órdenes LIMIT
3. **Revisa History** para analizar tus trades pasados
4. **Usa Ping** si sospechas problemas de conectividad
5. **Production Test** para probar recetas antes de producción real
6. **Keyboard shortcuts** hacen el trading más rápido
7. **Refresh Orders** antes de decisiones importantes

---

## 📞 Soporte

Para reportar bugs o solicitar features:
- GitHub Issues: https://github.com/sst/opencode/issues
- Comandos en CLI: `/help`

---

**Última actualización:** Noviembre 2025
**Versión:** 2.0 - Con todas las funcionalidades implementadas
