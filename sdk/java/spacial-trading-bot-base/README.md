# 🚀 Spacial Trading Bot Base

Cliente CLI para la Bolsa Interestelar de Aguacates Andorianos - Proyecto base para estudiantes.

## 📋 Tabla de Contenidos

- [Requisitos Previos](#requisitos-previos)
- [Configuración Inicial](#configuración-inicial)
- [Autenticación con GitHub Packages](#autenticación-con-github-packages)
- [Configuración de IntelliJ IDEA](#configuración-de-intellij-idea)
- [Compilación y Ejecución](#compilación-y-ejecución)
- [Configuración del Bot](#configuración-del-bot)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Herramientas de Calidad de Código](#herramientas-de-calidad-de-código)

---

## 🔧 Requisitos Previos

Antes de comenzar, asegúrate de tener instalado:

1. **Java 25** (JDK 25)
   - Descarga desde: https://jdk.java.net/25/
   - Verifica la instalación: `java -version`

2. **IntelliJ IDEA** (Community o Ultimate)
   - Descarga desde: https://www.jetbrains.com/idea/download/

3. **Git**
   - Descarga desde: https://git-scm.com/downloads
   - Verifica la instalación: `git --version`

4. **Cuenta de GitHub**
   - Necesaria para acceder al repositorio privado y al SDK

---

## ⚙️ Configuración Inicial

### 1. Clonar el Repositorio

Si el repositorio es privado, necesitarás permisos de acceso. Contacta al instructor para ser agregado al repositorio.

```bash
# Clonar usando HTTPS (te pedirá credenciales)
git clone https://github.com/HellSoft-Col/stock-market.git

# O usando SSH (requiere configurar llaves SSH)
git clone git@github.com:HellSoft-Col/stock-market.git

# Navegar al directorio del proyecto base
cd stock-market/sdk/java/spacial-trading-bot-base
```

### 2. Autenticación con GitHub Packages

El proyecto utiliza el SDK `websocket-client` que está alojado en GitHub Packages. Necesitas autenticarte para descargarlo.

#### 2.1 Generar un Personal Access Token (PAT)

1. Ve a GitHub → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Haz clic en **"Generate new token (classic)"**
3. Dale un nombre descriptivo (ej: "Trading Bot SDK Access")
4. Selecciona los siguientes scopes:
   - ✅ `read:packages` (obligatorio)
   - ✅ `repo` (si el repositorio es privado)
5. Haz clic en **"Generate token"**
6. **¡IMPORTANTE!** Copia el token inmediatamente (solo se muestra una vez)

#### 2.2 Configurar las Credenciales

Crea el archivo `gradle.properties` en la raíz del proyecto:

```bash
cp gradle.properties.sample gradle.properties
```

Edita `gradle.properties` y reemplaza los valores:

```properties
# GitHub Packages Authentication
gpr.user=TU_USUARIO_GITHUB
gpr.token=ghp_tu_token_aqui

# Gradle optimizations
org.gradle.daemon=true
org.gradle.caching=true
org.gradle.configuration-cache=true
```

**⚠️ IMPORTANTE:** El archivo `gradle.properties` está en `.gitignore` y **NO debe subirse a Git** porque contiene información sensible.

---

## 💻 Configuración de IntelliJ IDEA

### 1. Importar el Proyecto

1. Abre IntelliJ IDEA
2. Selecciona **"Open"** (no "New Project")
3. Navega hasta el directorio `spacial-trading-bot-base`
4. Selecciona el archivo `build.gradle.kts`
5. En el diálogo, selecciona **"Open as Project"**
6. IntelliJ detectará automáticamente que es un proyecto Gradle

### 2. Configurar el JDK 25

1. Ve a **File** → **Project Structure** (o `Cmd+;` en Mac, `Ctrl+Alt+Shift+S` en Windows/Linux)
2. En **"Project"**:
   - **SDK:** Selecciona o agrega Java 25
   - **Language level:** 25 (Preview)
3. Haz clic en **"OK"**

### 3. Sincronizar Gradle

IntelliJ sincronizará automáticamente las dependencias. Si no lo hace:

1. Abre el panel de **Gradle** (lado derecho de la ventana)
2. Haz clic en el ícono de **"Reload All Gradle Projects"** (🔄)

Si obtienes un error de autenticación:
- Verifica que `gradle.properties` exista y tenga las credenciales correctas
- Verifica que tu token de GitHub tenga el scope `read:packages`

### 4. Configurar Lombok (opcional)

El proyecto usa Lombok para reducir código repetitivo:

1. Ve a **File** → **Settings** → **Plugins**
2. Busca "Lombok" e instala el plugin
3. Reinicia IntelliJ
4. Ve a **Settings** → **Build, Execution, Deployment** → **Compiler** → **Annotation Processors**
5. Marca **"Enable annotation processing"**

### 5. Importar Configuración de Formato

El proyecto incluye configuración de formato de código:

1. Ve a **File** → **Settings** → **Editor** → **Code Style** → **Java**
2. Haz clic en el ícono de engranaje ⚙️ → **Import Scheme** → **Eclipse XML Profile**
3. Selecciona el archivo `config/eclipse-format.xml`
4. Haz clic en **"OK"**

---

## 🏗️ Compilación y Ejecución

### Usando IntelliJ IDEA

#### Compilar el Proyecto

1. Abre el panel de **Gradle** (lado derecho)
2. Navega a: **spacial-trading-bot-base** → **Tasks** → **build**
3. Doble clic en **"build"**

O desde el terminal integrado:
```bash
./gradlew build
```

#### Ejecutar el Programa

1. Abre la clase `tech.hellsoft.trading.Main`
2. Haz clic derecho en el archivo o en el método `main()`
3. Selecciona **"Run 'Main.main()'"**

O desde el terminal:
```bash
./gradlew run
```

### Usando la Terminal (Gradle)

```bash
# Compilar el proyecto
./gradlew build

# Compilar sin ejecutar tests
./gradlew build -x test

# Ejecutar el programa
./gradlew run

# Limpiar y compilar
./gradlew clean build

# Ejecutar tests
./gradlew test

# Ver todas las tareas disponibles
./gradlew tasks
```

---

## 📖 Entendiendo el Código de Ejemplo

El archivo `Main.java` contiene un ejemplo **simple y minimal** que muestra cómo conectarse al servidor de trading. Es un punto de partida para que implementes tu propia lógica.

### Estructura del Ejemplo

```java
public static void main(String[] args) {
    // 1️⃣ Cargar configuración (apiKey, team, host)
    Configuration config = ConfigLoader.load("src/main/resources/config.json");
    
    // 2️⃣ Crear conector y tu bot
    ConectorBolsa connector = new ConectorBolsa();
    MyTradingBot bot = new MyTradingBot();
    connector.addListener(bot);
    
    // 3️⃣ Conectar al servidor
    connector.conectar(config.host(), config.apiKey());
    
    // 4️⃣ Mantener el programa corriendo
    Thread.currentThread().join();
}
```

### Clase MyTradingBot (Tu Implementación)

El ejemplo incluye una clase interna `MyTradingBot` que implementa `EventListener`. Aquí es donde **tú implementarás tu estrategia de trading**:

#### Eventos Principales que Debes Manejar:

| Evento | Cuándo se Dispara | Qué Hacer |
|--------|-------------------|-----------|
| `onLoginOk()` | Conexión exitosa | Inicializar tu estado (balance, inventario inicial) |
| `onTicker()` | Actualización de precios | Decidir si comprar/vender basado en precios |
| `onFill()` | Orden ejecutada | Actualizar tu inventario y balance local |
| `onBalanceUpdate()` | Cambio en balance | Actualizar tu registro de dinero disponible |
| `onInventoryUpdate()` | Cambio en inventario | Actualizar tu registro de productos |
| `onError()` | Error del servidor | Manejar errores y reintentar si es necesario |

### Patrón "No Else" (Guard Clauses)

Nota cómo cada método usa **guard clauses** en lugar de `if-else`:

```java
@Override
public void onTicker(TickerMessage ticker) {
    // ✅ Guard clause: salir temprano si no hay datos
    if (ticker == null) {
        return;
    }
    
    // Lógica principal cuando ticker es válido
    System.out.println("Precio: " + ticker.getMid());
    
    // TODO: Tu estrategia de trading aquí
}
```

Este patrón es **obligatorio** según `AGENTS.md`. Evita anidación y hace el código más legible.

### ¿Qué Debes Implementar?

1. **Estado del Bot**: Agrega variables de instancia para rastrear:
   ```java
   private double balance;
   private Map<String, Integer> inventory;
   private Map<String, Double> prices;
   ```

2. **Lógica de Trading**: En `onTicker()`, implementa:
   - Detectar oportunidades de compra/venta
   - Calcular ganancias potenciales
   - Enviar órdenes usando el `ConectorBolsa`

3. **Producción**: Si tu rol permite producir:
   - Verifica ingredientes en `onInventoryUpdate()`
   - Calcula cuánto producir (algoritmo recursivo)
   - Envía comando de producción

4. **Gestión de Errores**: En `onError()`:
   - Registra errores
   - Implementa lógica de retry
   - Ajusta tu estrategia

### Ejemplo de Extensión (Para Estudiantes)

```java
private static class MyTradingBot implements EventListener {
    // Estado del bot
    private double currentBalance = 0;
    private Map<String, Integer> inventory = new HashMap<>();
    private Map<String, Double> lastPrices = new HashMap<>();
    
    @Override
    public void onLoginOk(LoginOKMessage loginOk) {
        if (loginOk == null) {
            return;
        }
        
        // Inicializar estado
        currentBalance = loginOk.getCurrentBalance();
        System.out.println("Balance inicial: $" + currentBalance);
    }
    
    @Override
    public void onTicker(TickerMessage ticker) {
        if (ticker == null) {
            return;
        }
        
        // Guardar precio
        lastPrices.put(ticker.getProduct(), ticker.getMid());
        
        // Estrategia simple: comprar si el precio es bajo
        if (ticker.getMid() < 50.0 && currentBalance > 100.0) {
            // TODO: Enviar orden de compra usando ConectorBolsa
            System.out.println("💡 Oportunidad de compra: " + ticker.getProduct());
        }
    }
    
    // ... otros métodos
}
```

### Siguientes Pasos

1. **Ejecuta el ejemplo** para ver cómo funciona
2. **Lee los eventos** que llegan del servidor
3. **Implementa tu estrategia** en los métodos TODO
4. **Consulta AGENTS.md** para patrones de diseño
5. **Agrega tests** para tu lógica

---

## 🤖 Configuración del Bot

### 1. Crear el Archivo de Configuración

El bot requiere un archivo `config.json` en `src/main/resources/`:

```bash
cp src/main/resources/config.sample.json src/main/resources/config.json
```

### 2. Editar la Configuración

Edita `src/main/resources/config.json`:

```json
{
  "apiKey": "TK-TU-TOKEN-AQUI",
  "team": "Nombre de tu Equipo",
  "host": "wss://trading.hellsoft.tech/ws"
}
```

**Dónde obtener tu API Key:**
- Tu instructor te proporcionará el token de acceso para el servidor de trading
- **NO compartas tu token** con otros equipos
- **NO subas `config.json` a Git** (está en `.gitignore`)

### 3. Configuración de Logging (Opcional)

El proyecto incluye `simplelogger.properties` para controlar los logs del SDK:

```properties
# src/main/resources/simplelogger.properties
org.slf4j.simpleLogger.defaultLogLevel=WARN
```

**Para ver más detalles del SDK** (útil para debugging), cambia a `INFO` o `DEBUG`:

```properties
org.slf4j.simpleLogger.defaultLogLevel=INFO
# O para debugging detallado:
# org.slf4j.simpleLogger.defaultLogLevel=DEBUG
```

---

## 📁 Estructura del Proyecto

### Código Fuente (Simplificado - Solo 4 archivos)

El proyecto base incluye **solo lo esencial** para que empieces:

```
src/main/java/tech/hellsoft/trading/
├── Main.java                        # 🚀 TU PUNTO DE PARTIDA
│                                    #    - Ejemplo simple de conexión
│                                    #    - Clase MyTradingBot con TODOs
│                                    #    - ¡Aquí implementas tu estrategia!
│
├── config/
│   └── Configuration.java           # Record con apiKey, team, host
│
├── exception/
│   └── ConfiguracionInvalidaException.java  # Errores de configuración
│
└── util/
    └── ConfigLoader.java            # Carga config.json
```

**¡Solo 4 archivos!** Todo lo demás lo crearás tú según necesites.

### Estructura Completa del Proyecto

```
spacial-trading-bot-base/
├── config/                          # Herramientas de calidad de código
│   ├── checkstyle/checkstyle.xml   # Reglas de estilo
│   ├── pmd/ruleset.xml              # Análisis estático
│   └── eclipse-format.xml           # Formato de código
│
├── gradle/wrapper/                  # Gradle wrapper (no tocar)
│
├── src/
│   └── main/
│       ├── java/                    # 👈 TU CÓDIGO AQUÍ (4 archivos base)
│       └── resources/
│           └── config.sample.json   # Plantilla de configuración
│
├── build.gradle.kts                 # Dependencias y plugins
├── settings.gradle.kts              # Configuración Gradle
├── gradle.properties.sample         # Plantilla (copiar y editar)
├── .java-version                    # Java 25
├── .gitignore                       # Archivos a ignorar
├── AGENTS.md                        # 📖 Guía de diseño (léela!)
└── README.md                        # Este archivo
```

### ¿Qué Archivos Crearás Tú?

Según `AGENTS.md`, probablemente necesitarás crear:

```
src/main/java/tech/hellsoft/trading/
├── model/
│   ├── Role.java                    # Datos de tu rol (especies, energía, etc.)
│   └── Recipe.java                  # Recetas de producción
│
├── exception/                       # Tus excepciones de negocio (7 mínimo)
│   ├── SaldoInsuficienteException.java
│   ├── InventarioInsuficienteException.java
│   ├── ProductoNoAutorizadoException.java
│   ├── IngredientesInsuficientesException.java
│   ├── RecetaNoEncontradaException.java
│   └── ...                          # Y más según necesites
│
└── strategy/                        # Tu lógica de trading
    ├── TradingStrategy.java
    ├── ProductionCalculator.java    # Algoritmo recursivo
    └── InventoryManager.java
```

**Principio clave**: Empieza simple, agrega complejidad solo cuando la necesites.

### Archivos que NO deben subirse a Git

Estos archivos están en `.gitignore` porque contienen información sensible o son generados automáticamente:

- `gradle.properties` - Credenciales de GitHub
- `src/main/resources/config.json` - Token de API del bot
- `build/` - Archivos compilados
- `.gradle/` - Cache de Gradle
- `.idea/workspace.xml` - Configuración personal de IntelliJ

---

## 🔍 Herramientas de Calidad de Código

El proyecto incluye tres herramientas de análisis de código:

### 1. Spotless (Formateo automático)

```bash
# Verificar el formato del código
./gradlew spotlessCheck

# Aplicar formato automáticamente
./gradlew spotlessApply
```

**Recomendación:** Ejecuta `spotlessApply` antes de cada commit.

### 2. Checkstyle (Estilo de código)

```bash
# Verificar el estilo de código
./gradlew checkstyleMain
./gradlew checkstyleTest

# Ver el reporte en:
# build/reports/checkstyle/main.html
```

### 3. PMD (Análisis estático)

```bash
# Ejecutar análisis estático
./gradlew pmdMain
./gradlew pmdTest

# Ver el reporte en:
# build/reports/pmd/main.html
```

### Verificar Todo

```bash
# Ejecutar todas las verificaciones + tests
./gradlew check

# Formatear y verificar
./gradlew spotlessApply check
```

---

## 🐛 Solución de Problemas Comunes

### Error: "Could not resolve tech.hellsoft.trading:websocket-client"

**Causa:** No se puede acceder a GitHub Packages.

**Solución:**
1. Verifica que `gradle.properties` existe y tiene las credenciales correctas
2. Verifica que tu token de GitHub tenga el scope `read:packages`
3. Prueba regenerar el token en GitHub
4. En IntelliJ: **Gradle** → **Reload All Gradle Projects**

### Error: "Unsupported class file major version 69"

**Causa:** Estás usando una versión de Java anterior a Java 25.

**Solución:**
1. Instala JDK 25
2. En IntelliJ: **File** → **Project Structure** → **Project** → **SDK:** Java 25
3. Reinicia IntelliJ

### El programa no encuentra config.json

**Causa:** No has creado el archivo de configuración.

**Solución:**
```bash
cp src/main/resources/config.sample.json src/main/resources/config.json
# Luego edita config.json con tu API key
```

### IntelliJ no reconoce las clases del SDK

**Causa:** Las dependencias no se descargaron correctamente.

**Solución:**
1. **File** → **Invalidate Caches** → **Invalidate and Restart**
2. Espera a que IntelliJ reconstruya el índice
3. Si persiste: elimina `.gradle/` y `.idea/`, luego reabre el proyecto

---

## 📚 Recursos Adicionales

- **Guía de desarrollo:** Lee `AGENTS.md` para entender los principios de diseño
- **SDK Documentation:** Consulta el Javadoc en GitHub Packages
- **Java 25 Features:** https://openjdk.org/projects/jdk/25/

---

## 📝 Notas Importantes

1. **NO subas archivos sensibles a Git:**
   - `gradle.properties` (credenciales de GitHub)
   - `config.json` (token de la API del bot)

2. **Antes de cada commit:**
   ```bash
   ./gradlew spotlessApply
   ./gradlew check
   ```

3. **Para trabajar en equipo:**
   - Cada miembro necesita su propio `gradle.properties`
   - Pueden compartir el mismo `config.json` (token del equipo)
   - Sincronicen cambios frecuentemente con Git

4. **Estilo de código:**
   - El proyecto sigue el principio **"No Else"**
   - Usa guard clauses, switch expressions, y patrones de diseño
   - Consulta `AGENTS.md` para detalles

---

## 🆘 Soporte

Si tienes problemas:

1. Revisa la sección de **Solución de Problemas** arriba
2. Consulta con tus compañeros de equipo
3. Busca en la documentación de Java 25
4. Contacta al instructor

---

**¡Buena suerte con tu bot de trading! 🚀🥑**
