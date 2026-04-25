package detector

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/raftweave/raftweave/internal/build/domain"
)

// Detector analyses a source directory and returns a DetectionResult.
type Detector interface {
	Detect(ctx context.Context, sourceDir string) (*domain.DetectionResult, error)
}

type compositeDetector struct{}

// New returns the default composite detector.
func New() Detector {
	return &compositeDetector{}
}

func (d *compositeDetector) Detect(ctx context.Context, sourceDir string) (*domain.DetectionResult, error) {
	// Empty directory check
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, domain.ErrDetectionFailed
	}
	if len(files) == 0 {
		return nil, domain.ErrDetectionFailed
	}

	// 1. Dockerfile
	if fileExists(filepath.Join(sourceDir, "Dockerfile")) || fileExists(filepath.Join(sourceDir, "dockerfile")) {
		return &domain.DetectionResult{
			HasDockerfile: true,
			Confidence:    1.0,
		}, nil
	}

	// 2. Go (go.mod)
	if fileExists(filepath.Join(sourceDir, "go.mod")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguageGo,
			Confidence:   0.5,
			ExposedPort:  8080,
			BuildCommand: "go build -o server .",
			StartCommand: "./server",
		}
		content, _ := os.ReadFile(filepath.Join(sourceDir, "go.mod"))
		contentStr := string(content)

		reVersion := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)
		if m := reVersion.FindStringSubmatch(contentStr); len(m) > 1 {
			res.RuntimeVersion = m[1]
			res.Confidence += 0.15
		}

		if strings.Contains(contentStr, "github.com/gin-gonic/gin") || strings.Contains(contentStr, "github.com/gin-gerson/gin") {
			res.Framework = "gin"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "github.com/labstack/echo") {
			res.Framework = "echo"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "github.com/gofiber/fiber") {
			res.Framework = "fiber"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "github.com/go-chi/chi") {
			res.Framework = "chi"
			res.Confidence += 0.2
		}

		return res, nil
	}

	// 3. Node (package.json)
	if fileExists(filepath.Join(sourceDir, "package.json")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguageNode,
			Confidence:   0.5,
			ExposedPort:  3000,
			BuildCommand: "npm install",
			StartCommand: "npm start",
		}
		content, _ := os.ReadFile(filepath.Join(sourceDir, "package.json"))
		contentStr := string(content)

		if strings.Contains(contentStr, `"express"`) {
			res.Framework = "express"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, `"next"`) {
			res.Framework = "next"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, `"@nestjs/core"`) {
			res.Framework = "nest"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, `"fastify"`) {
			res.Framework = "fastify"
			res.Confidence += 0.2
		}

		if fileExists(filepath.Join(sourceDir, ".nvmrc")) {
			nvmrc, _ := os.ReadFile(filepath.Join(sourceDir, ".nvmrc"))
			res.RuntimeVersion = strings.TrimSpace(string(nvmrc))
			res.Confidence += 0.15
		} else {
			reNodeVer := regexp.MustCompile(`"node":\s*"([^"]+)"`)
			if m := reNodeVer.FindStringSubmatch(contentStr); len(m) > 1 {
				res.RuntimeVersion = m[1]
				res.Confidence += 0.15
			}
		}

		return res, nil
	}

	// 4. Python (requirements.txt or pyproject.toml)
	if fileExists(filepath.Join(sourceDir, "requirements.txt")) || fileExists(filepath.Join(sourceDir, "pyproject.toml")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguagePython,
			Confidence:   0.5,
			ExposedPort:  8000,
			BuildCommand: "pip install -r requirements.txt",
			StartCommand: "python main.py",
		}
		
		var contentStr string
		if content, err := os.ReadFile(filepath.Join(sourceDir, "requirements.txt")); err == nil {
			contentStr += string(content)
		}
		if content, err := os.ReadFile(filepath.Join(sourceDir, "pyproject.toml")); err == nil {
			contentStr += string(content)
		}

		contentLower := strings.ToLower(contentStr)
		if strings.Contains(contentLower, "django") {
			res.Framework = "django"
			res.Confidence += 0.2
		} else if strings.Contains(contentLower, "flask") {
			res.Framework = "flask"
			res.Confidence += 0.2
		} else if strings.Contains(contentLower, "fastapi") {
			res.Framework = "fastapi"
			res.Confidence += 0.2
		}

		if fileExists(filepath.Join(sourceDir, ".python-version")) {
			pyVer, _ := os.ReadFile(filepath.Join(sourceDir, ".python-version"))
			res.RuntimeVersion = strings.TrimSpace(string(pyVer))
			res.Confidence += 0.15
		}

		return res, nil
	}

	// 5. Java (pom.xml or build.gradle)
	if fileExists(filepath.Join(sourceDir, "pom.xml")) || fileExists(filepath.Join(sourceDir, "build.gradle")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguageJava,
			Confidence:   0.5,
			ExposedPort:  8080,
			BuildCommand: "mvn package",
			StartCommand: "java -jar target/app.jar",
		}
		var contentStr string
		if content, err := os.ReadFile(filepath.Join(sourceDir, "pom.xml")); err == nil {
			contentStr += string(content)
		}
		if content, err := os.ReadFile(filepath.Join(sourceDir, "build.gradle")); err == nil {
			contentStr += string(content)
		}

		if strings.Contains(contentStr, "spring-boot") {
			res.Framework = "spring-boot"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "quarkus") {
			res.Framework = "quarkus"
			res.Confidence += 0.2
		}

		reJavaVer := regexp.MustCompile(`<java\.version>(.*?)</java\.version>`)
		if m := reJavaVer.FindStringSubmatch(contentStr); len(m) > 1 {
			res.RuntimeVersion = m[1]
			res.Confidence += 0.15
		}

		return res, nil
	}

	// 6. Ruby (Gemfile)
	if fileExists(filepath.Join(sourceDir, "Gemfile")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguageRuby,
			Confidence:   0.5,
			ExposedPort:  3000,
			BuildCommand: "bundle install",
			StartCommand: "bundle exec rails s",
		}
		content, _ := os.ReadFile(filepath.Join(sourceDir, "Gemfile"))
		contentStr := string(content)

		if strings.Contains(contentStr, "'rails'") || strings.Contains(contentStr, `"rails"`) {
			res.Framework = "rails"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "'sinatra'") || strings.Contains(contentStr, `"sinatra"`) {
			res.Framework = "sinatra"
			res.Confidence += 0.2
		}

		if fileExists(filepath.Join(sourceDir, ".ruby-version")) {
			rubyVer, _ := os.ReadFile(filepath.Join(sourceDir, ".ruby-version"))
			res.RuntimeVersion = strings.TrimSpace(string(rubyVer))
			res.Confidence += 0.15
		}

		return res, nil
	}

	// 7. Rust (Cargo.toml)
	if fileExists(filepath.Join(sourceDir, "Cargo.toml")) {
		res := &domain.DetectionResult{
			Language:     domain.LanguageRust,
			Confidence:   0.5,
			ExposedPort:  8080,
			BuildCommand: "cargo build --release",
			StartCommand: "./target/release/app",
		}
		content, _ := os.ReadFile(filepath.Join(sourceDir, "Cargo.toml"))
		contentStr := string(content)

		if strings.Contains(contentStr, "actix") {
			res.Framework = "actix"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "axum") {
			res.Framework = "axum"
			res.Confidence += 0.2
		} else if strings.Contains(contentStr, "warp") {
			res.Framework = "warp"
			res.Confidence += 0.2
		}

		reEdition := regexp.MustCompile(`edition\s*=\s*"([^"]+)"`)
		if m := reEdition.FindStringSubmatch(contentStr); len(m) > 1 {
			res.RuntimeVersion = m[1]
			res.Confidence += 0.15
		}

		return res, nil
	}

	// 8. .NET (*.csproj or *.sln)
	isDotnet := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".csproj") || strings.HasSuffix(f.Name(), ".sln") {
			isDotnet = true
			break
		}
	}
	if isDotnet {
		res := &domain.DetectionResult{
			Language:     domain.LanguageDotnet,
			Confidence:   0.5,
			ExposedPort:  5000,
			BuildCommand: "dotnet publish -c Release -o out",
			StartCommand: "dotnet out/app.dll",
		}
		
		// Find csproj
		var contentStr string
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".csproj") {
				content, _ := os.ReadFile(filepath.Join(sourceDir, f.Name()))
				contentStr += string(content)
			}
		}

		if strings.Contains(contentStr, "Microsoft.AspNetCore") || strings.Contains(contentStr, "Microsoft.NET.Sdk.Web") {
			res.Framework = "AspNetCore"
			res.Confidence += 0.2
		}

		reTargetFramework := regexp.MustCompile(`<TargetFramework>(.*?)</TargetFramework>`)
		if m := reTargetFramework.FindStringSubmatch(contentStr); len(m) > 1 {
			res.RuntimeVersion = m[1]
			res.Confidence += 0.15
		}

		return res, nil
	}

	return &domain.DetectionResult{
		Language:   domain.LanguageUnknown,
		Confidence: 0.1,
	}, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
