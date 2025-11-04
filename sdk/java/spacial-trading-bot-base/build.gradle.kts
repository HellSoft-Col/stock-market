plugins {
    java
    application
    id("com.diffplug.spotless") version "6.25.0"
}

group = "tech.hellsoft.trading"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
    maven {
        url = uri("https://maven.pkg.github.com/HellSoft-Col/stock-market")
        credentials {
            username = project.findProperty("gpr.user") as String? ?: System.getenv("GITHUB_ACTOR")
            password = project.findProperty("gpr.token") as String? ?: System.getenv("GITHUB_TOKEN")
        }
    }
}

dependencies {
    implementation("tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT")
    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

// Auto-formatting with Spotless
spotless {
    java {
        googleJavaFormat()
    }
}

// Task to auto-format code
tasks.register("formatAll") {
    group = "formatting"
    description = "Auto-format all Java code"
    dependsOn("spotlessApply")
}

// Task to check formatting
tasks.register("lintAll") {
    group = "verification"
    description = "Check code formatting"
    dependsOn("spotlessCheck")
}

tasks.test {
    useJUnitPlatform()
}

