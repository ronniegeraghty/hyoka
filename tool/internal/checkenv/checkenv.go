// Package checkenv checks for the presence of required tools and dependencies.
package checkenv

import (
	"fmt"
	"os/exec"
	"strings"
)

type checkResult struct {
	name    string
	ok      bool
	version string
	hint    string
}

// Run checks for language toolchains, Copilot CLI, and MCP prerequisites, then prints results.
func Run() {
	fmt.Println("Language Toolchains:")
	langChecks := []checkResult{
		checkPython(),
		checkDotnet(),
		checkGo(),
		checkNode(),
		checkJava(),
		checkRust(),
		checkCpp(),
	}
	for _, c := range langChecks {
		printCheck(c)
	}

	fmt.Println()
	fmt.Println("Copilot:")
	copilotChecks := []checkResult{
		checkCopilotCLI(),
		checkCopilotAuth(),
	}
	for _, c := range copilotChecks {
		printCheck(c)
	}

	fmt.Println()
	fmt.Println("MCP Servers:")
	mcpChecks := []checkResult{
		checkNpx(),
	}
	for _, c := range mcpChecks {
		printCheck(c)
	}
}

func printCheck(c checkResult) {
	if c.ok {
		fmt.Printf("  ✅ %-12s %s\n", c.name, c.version)
	} else {
		fmt.Printf("  ❌ %-12s %s\n", c.name, c.hint)
	}
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func extractVersion(output string) string {
	// Many tools output "toolname vX.Y.Z" or "toolname X.Y.Z" — grab first line, trim
	lines := strings.SplitN(output, "\n", 2)
	return strings.TrimSpace(lines[0])
}

func checkPython() checkResult {
	// Try python3 first, then python
	for _, bin := range []string{"python3", "python"} {
		if out, err := runCmd(bin, "--version"); err == nil {
			ver := extractVersion(out)
			// Also check pip
			pipVer := ""
			for _, pipBin := range []string{"pip3", "pip"} {
				if pout, perr := runCmd(pipBin, "--version"); perr == nil {
					parts := strings.Fields(pout)
					if len(parts) >= 2 {
						pipVer = parts[0] + " " + parts[1]
					}
					break
				}
			}
			display := ver
			if pipVer != "" {
				display += ", " + pipVer
			}
			return checkResult{name: "Python", ok: true, version: display}
		}
	}
	return checkResult{name: "Python", ok: false, hint: "not found (need: python3, pip)"}
}

func checkDotnet() checkResult {
	out, err := runCmd("dotnet", "--version")
	if err != nil {
		return checkResult{name: ".NET", ok: false, hint: "not found (need: dotnet)"}
	}
	return checkResult{name: ".NET", ok: true, version: "dotnet " + extractVersion(out)}
}

func checkGo() checkResult {
	out, err := runCmd("go", "version")
	if err != nil {
		return checkResult{name: "Go", ok: false, hint: "not found (need: go)"}
	}
	// "go version go1.26.1 linux/amd64" → extract version
	parts := strings.Fields(out)
	ver := out
	if len(parts) >= 3 {
		ver = parts[2] // "go1.26.1"
	}
	return checkResult{name: "Go", ok: true, version: ver}
}

func checkNode() checkResult {
	nodeOut, err := runCmd("node", "--version")
	if err != nil {
		return checkResult{name: "Node.js", ok: false, hint: "not found (need: node, npm)"}
	}
	ver := "node " + extractVersion(nodeOut)
	if npmOut, nerr := runCmd("npm", "--version"); nerr == nil {
		ver += ", npm " + extractVersion(npmOut)
	}
	return checkResult{name: "Node.js", ok: true, version: ver}
}

func checkJava() checkResult {
	out, err := runCmd("java", "-version")
	if err != nil {
		// java -version writes to stderr on some versions; try --version
		out, err = runCmd("java", "--version")
	}
	if err != nil {
		return checkResult{name: "Java", ok: false, hint: "not found (need: java, mvn or gradle)"}
	}
	ver := extractVersion(out)
	// Check for build tools
	tools := []string{}
	if _, merr := runCmd("mvn", "--version"); merr == nil {
		tools = append(tools, "mvn")
	}
	if _, gerr := runCmd("gradle", "--version"); gerr == nil {
		tools = append(tools, "gradle")
	}
	if len(tools) > 0 {
		ver += " (" + strings.Join(tools, ", ") + ")"
	}
	return checkResult{name: "Java", ok: true, version: ver}
}

func checkRust() checkResult {
	out, err := runCmd("cargo", "--version")
	if err != nil {
		return checkResult{name: "Rust", ok: false, hint: "not found (need: cargo)"}
	}
	return checkResult{name: "Rust", ok: true, version: extractVersion(out)}
}

func checkCpp() checkResult {
	versions := []string{}
	for _, compiler := range []string{"gcc", "g++", "clang"} {
		if out, err := runCmd(compiler, "--version"); err == nil {
			versions = append(versions, compiler+" "+extractVersion(out))
			break // one compiler is enough
		}
	}
	if cmakeOut, err := runCmd("cmake", "--version"); err == nil {
		line := extractVersion(cmakeOut)
		// "cmake version 3.28.1" → extract
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			versions = append(versions, "cmake "+parts[2])
		} else {
			versions = append(versions, line)
		}
	}
	if len(versions) == 0 {
		return checkResult{name: "C/C++", ok: false, hint: "not found (need: gcc or clang, cmake)"}
	}
	return checkResult{name: "C/C++", ok: true, version: strings.Join(versions, ", ")}
}

func checkCopilotCLI() checkResult {
	out, err := runCmd("copilot", "--version")
	if err != nil {
		return checkResult{name: "Copilot CLI", ok: false, hint: "not found (install: https://docs.github.com/en/copilot)"}
	}
	return checkResult{name: "Copilot CLI", ok: true, version: extractVersion(out)}
}

func checkCopilotAuth() checkResult {
	out, err := runCmd("gh", "auth", "status")
	if err != nil {
		return checkResult{name: "Authenticated", ok: false, hint: "not authenticated (run: gh auth login)"}
	}
	// Try to extract username from output
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Logged in to") || strings.Contains(line, "account") {
			return checkResult{name: "Authenticated", ok: true, version: strings.TrimSpace(line)}
		}
	}
	return checkResult{name: "Authenticated", ok: true, version: "authenticated via gh"}
}

func checkNpx() checkResult {
	out, err := runCmd("npx", "--version")
	if err != nil {
		return checkResult{name: "npx", ok: false, hint: "not found (need: npx for Azure MCP)"}
	}
	return checkResult{name: "npx", ok: true, version: "npx " + extractVersion(out) + " (for Azure MCP)"}
}
