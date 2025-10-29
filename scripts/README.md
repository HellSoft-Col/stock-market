# Trading Server Simulation Scripts

Este directorio contiene scripts de simulación para probar el servidor de trading con múltiples clientes concurrentes, validando órdenes, producción en ráfaga y condiciones competitivas.

## 📋 Descripción

El sistema de simulación crea múltiples clientes WebSocket que:
- Se conectan con diferentes tokens de equipo
- Realizan producción en ráfaga durante los primeros minutos
- Ejecutan órdenes de compra/venta competitivas
- Validan que el servidor procese las órdenes por orden de llegada
- Monitorizan tiempos de respuesta y estadísticas de rendimiento

## 🚀 Instalación y Configuración

### Requisitos
- Python 3.8+
- Servidor de trading ejecutándose (por defecto en `ws://localhost:8080`)

### Instalación automática de dependencias
```bash
# Opción 1: Instalación automática
python3 run_simulation.py --install-deps

# Opción 2: Instalación manual
pip install websockets
```

## 📖 Uso

### Opción 1: Usando el script runner (Recomendado)
```bash
# Simulación básica con tokens específicos
python3 run_simulation.py --tokens TK-1001,TK-1002,TK-1003 --duration 15

# Usando archivo de configuración
python3 run_simulation.py --config simulation_config.json

# Modo verbose para debugging
python3 run_simulation.py --tokens TK-1001,TK-1002,TK-1003 --verbose
```

### Opción 2: Script directo
```bash
# Después de instalar websockets manualmente
python3 trading_simulation.py --tokens TK-1001,TK-1002,TK-1003 --duration 15
```

## ⚙️ Configuración

### Archivo de configuración (`simulation_config.json`)
```json
{
  "simulation": {
    "duration_minutes": 15,
    "server_url": "ws://localhost:8080",
    "log_level": "INFO"
  },
  "tokens": [
    "TK-1001",
    "TK-1002", 
    "TK-1003",
    "TK-1004",
    "TK-1005"
  ],
  "phases": {
    "burst_production": {
      "duration_minutes": 2
    },
    "mixed_trading": {
      "duration_minutes": 10
    },
    "competitive_trading": {
      "duration_minutes": 3
    }
  }
}
```

## 🏗️ Fases de Simulación

### Fase 1: Producción en Ráfaga (2 minutos)
- **Objetivo:** Crear inventario inicial para trading
- **Actividades:**
  - Producción intensiva de productos básicos (FOSFO, PITA, PALTA-OIL)
  - Cantidades aleatorias entre 10-30 unidades
  - Intervalos de 5-15 segundos entre producciones

### Fase 2: Trading Mixto (10 minutos)
- **Objetivo:** Simular actividad normal de mercado
- **Actividades:**
  - 30% órdenes de compra
  - 30% órdenes de venta
  - 20% producción adicional
  - 20% períodos de espera
  - Precios realistas con variaciones del mercado

### Fase 3: Trading Competitivo (3 minutos)
- **Objetivo:** Probar validación de órdenes first-come-first-served
- **Actividades:**
  - Órdenes agresivas con precios competitivos
  - Múltiples clientes compitiendo por las mismas órdenes
  - Validación de prioridad temporal en el servidor

## 📊 Métricas y Validación

### Lo que se prueba:
1. **Conectividad WebSocket:** Múltiples conexiones concurrentes
2. **Autenticación:** Validación de tokens de equipo
3. **Órdenes First-Come-First-Served:** El servidor acepta la primera orden válida
4. **Validación de Fondos:** Verificación de balance e inventario
5. **Producción:** Algoritmos de producción y actualización de inventario
6. **Rendimiento:** Tiempos de respuesta y throughput

### Estadísticas reportadas:
- Órdenes totales vs exitosas vs fallidas
- Tiempo promedio de respuesta
- Producciones completadas por cliente
- Fills (ejecuciones) recibidas
- Estadísticas por cliente individual

## 📝 Logs y Debugging

### Archivos de log
Los logs se guardan automáticamente en:
```
trading_simulation_YYYYMMDD_HHMMSS.log
```

### Levels de logging
- `INFO`: Información general de la simulación
- `DEBUG`: Detalles de mensajes WebSocket y timing
- `WARNING`: Errores recuperables
- `ERROR`: Errores críticos

### Ejemplo de output
```
2024-01-15 10:30:00 - Simulation - INFO - Starting 15-minute trading simulation with 3 clients
2024-01-15 10:30:01 - Client-TK-1001 - INFO - Successfully authenticated with token TK-1001
2024-01-15 10:30:02 - Client-TK-1001 - INFO - Production completed: 15 FOSFO
2024-01-15 10:30:05 - Client-TK-1002 - INFO - Order placed: BUY 5 PITA @ $18.50
2024-01-15 10:30:06 - Client-TK-1003 - INFO - Order filled: SELL 5 PITA @ $18.50
```

## 🔧 Troubleshooting

### Problemas comunes:

**Error: "Connection refused"**
```bash
# Verifica que el servidor esté ejecutándose
netstat -an | grep 8080
```

**Error: "Authentication failed"**
- Verifica que los tokens sean válidos
- Asegúrate de que el formato sea TK-XXXX

**Error: "websockets not found"**
```bash
pip install websockets
# o usar el runner con --install-deps
```

**Performance issues**
- Reduce el número de clientes concurrentes
- Aumenta los intervalos entre órdenes
- Verifica recursos del servidor

## 🎯 Casos de Uso

### Testing de Desarrollo
```bash
# Test rápido con 2 clientes por 5 minutos
python3 run_simulation.py --tokens TK-TEST1,TK-TEST2 --duration 5
```

### Testing de Stress
```bash
# Test intensivo con 10 clientes por 15 minutos
python3 run_simulation.py --tokens TK-1001,TK-1002,TK-1003,TK-1004,TK-1005,TK-1006,TK-1007,TK-1008,TK-1009,TK-1010 --duration 15
```

### Testing de Validación
```bash
# Test específico para validar order priority
python3 run_simulation.py --config validation_config.json --verbose
```

## 📈 Interpretación de Resultados

### Resultados exitosos:
- Tasa de éxito de órdenes > 95%
- Tiempo de respuesta promedio < 100ms
- Sin errores de conexión
- Distribución equitativa de fills entre clientes

### Indicadores de problemas:
- Alta tasa de órdenes fallidas
- Tiempos de respuesta > 1 segundo
- Desconexiones frecuentes
- Un cliente acaparando todos los fills

## 🤝 Contribución

Para añadir nuevas funcionalidades:
1. Modifica `trading_simulation.py` para nuevos tipos de órdenes
2. Actualiza `simulation_config.json` para nuevos parámetros
3. Añade tests específicos en las fases de simulación
4. Documenta los cambios en este README

## 📄 Licencia

Este código está incluido como parte del proyecto de trading de aguacates para fines educativos y de testing.