package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	index              int
	docroot            string
	eslint             bool
	prettier           bool
	stylelint          bool
	secretlint         bool
	phpcs              bool
	validateBranchName bool
	jiraPrepareCommit  bool
}

var questions = []string{
	"Give the path of your docroot (auto-detect if current directory has 'docroot' or 'web' folder): ",
	"Do you want to add eslint for JS? (y/n): ",
	"Do you want to add prettier support? (y/n): ",
	"Do you want to add stylelint for CSS and SCSS? (y/n): ",
	"Do you want to add secretlint for all files? (y/n): ",
	"Do you want to add PHPCS and PHPCBF for PHP and all Drupal PHP files? (y/n): ",
	"Do you want to add support for validating branch name pattern using validate-branch-name npm package? (y/n): ",
	"Do you want to add support to automatically add ticket number in commit message using jira-prepare-commit-msg npm package? (y/n): ",
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.index == 0 {
				if m.docroot == "" {
					if _, err := os.Stat("docroot"); err == nil {
						m.docroot = "docroot"
					} else if _, err := os.Stat("web"); err == nil {
						m.docroot = "web"
					} else {
						m.docroot = "."
					}
				}
			} else {
				answer := strings.ToLower(strings.TrimSpace(msg.String()))
				switch m.index {
				case 1:
					m.eslint = (answer == "y")
				case 2:
					m.prettier = (answer == "y")
				case 3:
					m.stylelint = (answer == "y")
				case 4:
					m.secretlint = (answer == "y")
				case 5:
					m.phpcs = (answer == "y")
				case 6:
					m.validateBranchName = (answer == "y")
				case 7:
					m.jiraPrepareCommit = (answer == "y")
				}
			}
			m.index++
			if m.index >= len(questions) {
				setupGitHooks(m)
				return m, tea.Quit
			}
		default:
			if m.index == 0 {
				m.docroot += msg.String()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.index >= len(questions) {
		return "Setting up your Git pre-commit hooks...\n"
	}
	return questions[m.index]
}

func setupGitHooks(m model) {
	fmt.Println("Setting up Git pre-commit hooks...")
	installPackages := []string{"husky", "lint-staged"}
	if m.eslint {
		installPackages = append(installPackages, "eslint")
		writeFile(".eslintrc.js", "module.exports = {\n  // ESLint configuration\n};\n")
	}
	if m.prettier {
		installPackages = append(installPackages, "prettier")
		writeFile(".prettierrc.js", "module.exports = {\n  // Prettier configuration\n};\n")
	}
	if m.stylelint {
		installPackages = append(installPackages, "stylelint")
		writeFile(".stylelintrc.js", "module.exports = {\n  // Stylelint configuration\n};\n")
	}
	if m.secretlint {
		installPackages = append(installPackages, "secretlint")
		writeFile(".secretlintrc.js", "module.exports = {\n  // Secretlint configuration\n};\n")
	}
	if m.phpcs {
		installPackages = append(installPackages, "phpcs")
		writeFile("phpcs.xml", "<ruleset name=\"Drupal\">\n  <!-- PHPCS configuration for Drupal -->\n</ruleset>\n")
	}
	if m.validateBranchName {
		installPackages = append(installPackages, "validate-branch-name")
		writeFile(".validate-branch-namerc.js", "module.exports = {\n  // validate-branch-name configuration\n};\n")
	}
	if m.jiraPrepareCommit {
		installPackages = append(installPackages, "jira-prepare-commit-msg")
		writeFile(".prepare-commit-msg", "#!/bin/sh\n# Script to automatically add ticket number to commit message\n")
	}
	runCommand("npm", append([]string{"install", "--save-dev"}, installPackages...)...)
	runCommand("npx", "husky", "install")
	runCommand("npx", "husky", "add", ".husky/pre-commit", "npx lint-staged")
	writeFile(".lintstagedrc.js", generateLintStagedConfig(m))
}

func generateLintStagedConfig(m model) string {
	config := "module.exports = {\n"
	if m.eslint {
		config += "  '*.js': ['eslint --fix'],\n"
	}
	if m.prettier {
		config += "  '*.js': ['prettier --write'],\n"
	}
	if m.stylelint {
		config += "  '*.css': ['stylelint --fix'],\n"
	}
	if m.secretlint {
		config += "  '*.*': ['secretlint'],\n"
	}
	if m.phpcs {
		config += "  '*.php': ['phpcs --standard=phpcs.xml'],\n"
	}
	config += "};\n"
	return config
}

func writeFile(filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
		os.Exit(1)
	}
}
