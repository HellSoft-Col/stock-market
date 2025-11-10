# Verificación de Setup para Estudiantes

## ✅ Checklist de Verificación Post-Clone

Después de clonar el repositorio, verifica que tengas estos archivos:

### 1. Archivos de Gradle (Esenciales para compilar)
```bash
ls -la gradlew gradlew.bat
ls -la gradle/wrapper/gradle-wrapper.jar
ls -la gradle/wrapper/gradle-wrapper.properties
ls -la build.gradle.kts settings.gradle.kts
```

### 2. Plantillas de Configuración
```bash
ls -la gradle.properties.sample
ls -la src/main/resources/config.sample.json
```

### 3. Configuración de Calidad de Código
```bash
ls -la config/eclipse-format.xml
ls -la config/checkstyle/checkstyle.xml
ls -la config/pmd/ruleset.xml
```

### 4. Documentación
```bash
ls -la README.md AGENTS.md
```

## 🔧 Pasos de Configuración Obligatorios

### Paso 1: Crear gradle.properties
```bash
# Desde la raíz del proyecto
cp gradle.properties.sample gradle.properties

# Editar y agregar tus credenciales de GitHub
nano gradle.properties  # o usa tu editor favorito
```

Contenido esperado de `gradle.properties`:
```properties
gpr.user=TU_USUARIO_GITHUB
gpr.token=ghp_TU_TOKEN_AQUI
org.gradle.daemon=true
org.gradle.caching=true
org.gradle.configuration-cache=true
```

### Paso 2: Crear config.json
```bash
# Desde la raíz del proyecto
cp src/main/resources/config.sample.json src/main/resources/config.json

# Editar y agregar tu API key del bot
nano src/main/resources/config.json
```

Contenido esperado de `config.json`:
```json
{
  "apiKey": "TK-TU-TOKEN-AQUI",
  "team": "Nombre de tu Equipo",
  "host": "wss://trading.hellsoft.tech/ws"
}
```

### Paso 3: Verificar que NO estén en Git
```bash
# Estos comandos NO deben mostrar los archivos
git status | grep gradle.properties
git status | grep config.json

# Si aparecen, están en git por error (¡MAL!)
```

## 🧪 Prueba de Compilación

Después de crear los archivos de configuración:

```bash
# 1. Compilar el proyecto (descarga dependencias)
./gradlew build

# Si falla con error de autenticación:
# - Verifica gradle.properties
# - Verifica que el token tenga scope 'read:packages'
# - Regenera el token en GitHub si es necesario

# 2. Ejecutar el programa (prueba básica)
./gradlew run

# Debería conectarse al servidor (si está en línea)
# o mostrar error de conexión (si está offline)
```

## 🐛 Solución Rápida de Problemas

### Error: "401 Unauthorized" al descargar dependencias
```bash
# Causa: Credenciales incorrectas en gradle.properties

# Solución:
# 1. Verifica que gradle.properties existe
cat gradle.properties

# 2. Verifica tu token en GitHub
# Settings → Developer settings → Personal access tokens
# El token debe tener scope: read:packages

# 3. Regenera gradle.properties
cp gradle.properties.sample gradle.properties
# Edita con las credenciales correctas
```

### Error: "config.json not found"
```bash
# Causa: No has creado el archivo de configuración

# Solución:
cp src/main/resources/config.sample.json src/main/resources/config.json
# Edita con tu API key
```

### Error: "Java version mismatch"
```bash
# Causa: No estás usando Java 25

# Verificar versión:
java -version

# Debe mostrar: java version "25" o "25-ea"

# Si no, descarga Java 25:
# https://jdk.java.net/25/
```

## 📊 Estado Esperado de Git

### Archivos que VES en el proyecto pero NO en Git:
```bash
gradle.properties              # Credenciales (ignorado)
src/main/resources/config.json # Token del bot (ignorado)
build/                         # Compilados (ignorado)
.gradle/                       # Cache (ignorado)
.idea/workspace.xml            # Config personal (ignorado)
```

### Verificar .gitignore:
```bash
cat .gitignore | grep -E "gradle.properties|config.json"

# Debe mostrar:
# gradle.properties
# (config.json está implícitamente ignorado con el patrón de resources)
```

## ✅ Checklist Final

- [ ] Clonado el repositorio
- [ ] Creado `gradle.properties` con credenciales de GitHub
- [ ] Creado `config.json` con API key del bot
- [ ] Ejecutado `./gradlew build` exitosamente
- [ ] Verificado que archivos sensibles NO están en git
- [ ] Importado el proyecto en IntelliJ
- [ ] Configurado JDK 25 en IntelliJ
- [ ] Instalado plugin de Lombok (opcional)
- [ ] Importado formato de código (eclipse-format.xml)
- [ ] Ejecutado `./gradlew run` para probar

## 📞 Contacto

Si después de seguir estos pasos aún tienes problemas:

1. Revisa el README.md completo
2. Consulta AGENTS.md para guía de desarrollo
3. Busca en los issues del repositorio
4. Contacta al instructor

---

**Nota:** Este documento es solo para verificación. Los estudiantes deben seguir el README.md principal para instrucciones completas.
