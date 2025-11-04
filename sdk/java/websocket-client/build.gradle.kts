plugins {
    `java-library`
    `maven-publish`
    jacoco
    checkstyle
    id("com.diffplug.spotless") version "6.25.0"
}

group = "tech.hellsoft.trading"
version = "1.0.0-SNAPSHOT"

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
        addStringOption("Xdoclint:none", "-quiet")
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

        // Use Google Java Format
        googleJavaFormat("1.19.2")

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

// Run spotless check before compilation
tasks.withType<JavaCompile> {
    dependsOn("spotlessApply")
}

// Add quality check task
tasks.register("qualityCheck") {
    group = "verification"
    description = "Run all quality checks: tests, coverage, formatting, and linting"
    dependsOn("clean", "spotlessCheck", "checkstyleMain", "checkstyleTest", "test", "jacocoTestReport")
}
