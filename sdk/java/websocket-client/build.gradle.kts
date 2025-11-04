plugins {
    `java-library`
    `maven-publish`
    jacoco
    checkstyle
    id("com.diffplug.spotless") version "6.25.0"
}

group = "tech.hellsoft.trading"

// Read version from VERSION file, fallback to default, allow override
val versionFile = file("VERSION")
val baseVersion = if (versionFile.exists()) {
    versionFile.readText().trim()
} else {
    "1.0.0-SNAPSHOT"
}

// Allow version override via command line (for CI builds)
version = project.findProperty("version") as String? ?: baseVersion

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
    withSourcesJar()
    withJavadocJar()
}

repositories {
    mavenCentral()
}

dependencies {
    // JSON Processing
    api("com.google.code.gson:gson:2.13.1")

    // Lombok - Code generation
    compileOnly("org.projectlombok:lombok:1.18.40")
    annotationProcessor("org.projectlombok:lombok:1.18.40")

    // Logging
    implementation("org.slf4j:slf4j-api:2.0.16")

    // Testing
    testImplementation(platform("org.junit:junit-bom:5.11.4"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testImplementation("org.junit.jupiter:junit-jupiter-params")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
    testImplementation("org.mockito:mockito-core:5.18.0")
    testImplementation("org.mockito:mockito-junit-jupiter:5.18.0")
    testCompileOnly("org.projectlombok:lombok:1.18.40")
    testAnnotationProcessor("org.projectlombok:lombok:1.18.40")
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)
}

// Ensure JavaDoc is generated during build
tasks.named("build") {
    dependsOn(tasks.javadoc)
}

// Ensure JavaDoc is included in check
tasks.named("check") {
    dependsOn(tasks.javadoc)
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required = true
        html.required = true
        csv.required = false
    }
}

tasks.jacocoTestCoverageVerification {
    violationRules {
        rule {
            limit {
                minimum = "0.80".toBigDecimal()
            }
        }
    }
}

tasks.withType<JavaCompile> {
    options.compilerArgs.addAll(
        listOf(
            "-Xlint:all",
            "-Xlint:-processing",
            "-Werror",
        ),
    )
}

tasks.javadoc {
    options {
        this as StandardJavadocDocletOptions
        // Configure doclint to be less strict but still catch important issues
        addStringOption("Xdoclint:all,-missing", "-quiet")
        // Add links to external documentation
        (this as StandardJavadocDocletOptions).links("https://docs.oracle.com/en/java/javase/25/docs/api/")
        // Set window title
        windowTitle = "Stock Market WebSocket Client API"
        // Set document title
        docTitle = "Stock Market WebSocket Client API Documentation"
        // Set header
        header = "Stock Market WebSocket Client"
        // Include author tags
        addBooleanOption("author", true)
        // Include version tags
        addBooleanOption("version", true)
        // Use HTML5
        addBooleanOption("html5", true)
        // Show protected members
        addBooleanOption("protected", true)
        // Show package private members
        addBooleanOption("package", true)
    }
}

tasks.withType<JavaExec> {
    jvmArgs =
        listOf(
            "--add-opens",
            "java.base/java.lang=ALL-UNNAMED",
            "--add-opens",
            "java.base/java.lang.invoke=ALL-UNNAMED",
        )
}

// Spotless: Code formatting
spotless {
    java {
        target("src/**/*.java")

        // Use Google Java Format (1.24.0 supports Java 25)
        googleJavaFormat("1.24.0")

        // Remove unused imports
        removeUnusedImports()

        // Format imports
        importOrder("java", "javax", "org", "com", "tech", "")

        // Ensure newline at end of file
        endWithNewline()

        // Trim trailing whitespace
        trimTrailingWhitespace()

        // Spotless doesn't need custom wildcard check - removeUnusedImports handles it
    }

    kotlinGradle {
        target("*.gradle.kts")
        ktlint()
    }
}

// Checkstyle: Linting and code quality
checkstyle {
    toolVersion = "9.3"
    configFile = file("$rootDir/config/checkstyle/checkstyle.xml")
    configProperties = mapOf("suppressionFile" to "$rootDir/config/checkstyle/suppressions.xml")
    isIgnoreFailures = true // Don't fail build on warnings
    maxWarnings = 100 // Allow some warnings
}

tasks.named<Checkstyle>("checkstyleMain") {
    reports {
        xml.required.set(true)
        html.required.set(true)
    }
}

tasks.named<Checkstyle>("checkstyleTest") {
    reports {
        xml.required.set(true)
        html.required.set(true)
    }
}

// Run spotless check before compilation (only in local development)
if (!System.getenv().containsKey("CI")) {
    tasks.withType<JavaCompile> {
        dependsOn("spotlessApply")
    }
}

// Add quality check task
tasks.register("qualityCheck") {
    group = "verification"
    description = "Run all quality checks: tests, coverage, formatting, linting, and documentation"
    dependsOn("clean", "spotlessCheck", "checkstyleMain", "checkstyleTest", "test", "jacocoTestReport", "javadoc")
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])

            pom {
                name.set("Stock Market WebSocket Client")
                description.set("Java WebSocket client library for connecting to the stock market trading system")
                url.set("https://github.com/${System.getenv("GITHUB_REPOSITORY") ?: "unknown/unknown"}")

                licenses {
                    license {
                        name.set("MIT License")
                        url.set("https://opensource.org/licenses/MIT")
                    }
                }

                developers {
                    developer {
                        id.set("hellsoft")
                        name.set("Hellsoft Tech")
                    }
                }

                scm {
                    connection.set("scm:git:git://github.com/${System.getenv("GITHUB_REPOSITORY") ?: "unknown/unknown"}.git")
                    developerConnection.set("scm:git:ssh://github.com:${System.getenv("GITHUB_REPOSITORY") ?: "unknown/unknown"}.git")
                    url.set("https://github.com/${System.getenv("GITHUB_REPOSITORY") ?: "unknown/unknown"}")
                }
            }
        }
    }

    repositories {
        maven {
            name = "GitHubPackages"
            url = uri("https://maven.pkg.github.com/${System.getenv("GITHUB_REPOSITORY") ?: "unknown/unknown"}")
            credentials {
                username = System.getenv("GITHUB_ACTOR")
                password = System.getenv("GITHUB_TOKEN")
            }
        }
    }
}
