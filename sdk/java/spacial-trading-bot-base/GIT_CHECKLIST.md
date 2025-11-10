# Lista de Verificación - Archivos para Git

## ✅ Archivos que DEBEN estar en Git

### Configuración del Proyecto
- [x] `build.gradle.kts` - Configuración de dependencias y build
- [x] `settings.gradle.kts` - Configuración de Gradle
- [x] `gradle.properties.sample` - Plantilla de propiedades (SIN credenciales)
- [x] `gradlew` - Gradle Wrapper (Linux/Mac)
- [x] `gradlew.bat` - Gradle Wrapper (Windows)
- [x] `gradle/wrapper/gradle-wrapper.jar` - JAR del wrapper
- [x] `gradle/wrapper/gradle-wrapper.properties` - Propiedades del wrapper

### Configuración de Herramientas
- [x] `.gitignore` - Archivos a ignorar
- [x] `.editorconfig` - Configuración de editor
- [x] `config/eclipse-format.xml` - Formato de código
- [x] `config/checkstyle/checkstyle.xml` - Reglas de Checkstyle
- [x] `config/pmd/ruleset.xml` - Reglas de PMD

### Código Fuente
- [x] `src/main/java/**/*.java` - Todo el código Java
- [x] `src/test/java/**/*.java` - Todos los tests
- [x] `src/main/resources/config.sample.json` - Plantilla de configuración (SIN tokens)

### Documentación
- [x] `README.md` - Instrucciones de setup
- [x] `AGENTS.md` - Guía de desarrollo

### Configuración de IntelliJ (compartida)
- [x] `.idea/codeStyles/` - Estilos de código compartidos

## ❌ Archivos que NO DEBEN estar en Git

### Archivos Sensibles (contienen tokens/credenciales)
- [ ] `gradle.properties` - Contiene credenciales de GitHub Packages
- [ ] `src/main/resources/config.json` - Contiene API key del bot

### Archivos Generados
- [ ] `build/` - Archivos compilados
- [ ] `.gradle/` - Cache de Gradle
- [ ] `bin/` - Archivos compilados (Eclipse)
- [ ] `out/` - Archivos compilados (IntelliJ)

### Configuración Personal de IDE
- [ ] `.idea/workspace.xml` - Configuración personal
- [ ] `.idea/misc.xml` - Configuración personal
- [ ] `.idea/modules.xml` - Generado automáticamente
- [ ] `*.iml` - Archivos de módulo de IntelliJ
- [ ] `.classpath` - Configuración de Eclipse
- [ ] `.project` - Configuración de Eclipse
- [ ] `.settings/` - Configuración de Eclipse

### Otros
- [ ] `.DS_Store` - Archivos de macOS
- [ ] `*.class` - Archivos compilados
- [ ] `*.jar` (excepto gradle-wrapper.jar)

## 🔍 Verificar antes de Commit

```bash
# Ver archivos que se van a commitear
git status

# Ver el contenido de los archivos staged
git diff --staged

# Verificar que no hay tokens/credenciales
git grep -i "token\|password\|apikey\|secret" -- ':!*.sample.*' ':!*.md'

# Formatear código antes de commit
./gradlew spotlessApply

# Verificar calidad de código
./gradlew check
```

## 📋 Comandos Útiles

```bash
# Ver archivos ignorados por git
git status --ignored

# Eliminar un archivo que fue commiteado por error
git rm --cached archivo.txt

# Si accidentalmente commiteaste credenciales:
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch gradle.properties" \
  --prune-empty --tag-name-filter cat -- --all
```

## ⚠️ IMPORTANTE

1. **NUNCA** subas `gradle.properties` con credenciales reales
2. **NUNCA** subas `config.json` con API keys reales
3. **SIEMPRE** usa los archivos `.sample` como referencia
4. **VERIFICA** con `git diff` antes de hacer commit
5. Si tienes dudas, pregunta antes de hacer push

## 🎯 Workflow Recomendado

1. Haz tus cambios en el código
2. Ejecuta `./gradlew spotlessApply`
3. Ejecuta `./gradlew check` para verificar
4. Revisa los cambios: `git status` y `git diff`
5. Agrega archivos: `git add <archivos>`
6. Verifica de nuevo: `git status`
7. Haz commit: `git commit -m "mensaje descriptivo"`
8. Push: `git push`

