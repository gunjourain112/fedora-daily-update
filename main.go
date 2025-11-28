package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle  = lipgloss.NewStyle().MarginLeft(1).MarginRight(5).Padding(0, 1).Italic(true).Foreground(lipgloss.Color("#FFF7DB")).SetString("System Updater")
	checkMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	crossMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("✗")
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
)

type model struct {
	tasks        []Task
	currentTask  int
	width        int
	height       int
	spinner      spinner.Model
	progress     progress.Model
	viewport     viewport.Model
	done         bool
	err          error
	outputChan   chan string
	finishedChan chan error
}

// Messages
type taskOutputMsg string
type taskFinishedMsg error

func initialModel() model {
	// Define tasks here
	tasks := []Task{
		{Name: "System Update (dnf)", Command: "sudo", Args: []string{"dnf", "update", "-y"}},
		{Name: "NPM Global Update", Command: "npm", Args: []string{"-g", "update"}},
		{Name: "Kiro-CLI Update", Command: "kiro-cli", Args: []string{"update"}},
		{Name: "Flatpak Update", Command: "flatpak", Args: []string{"update", "-y"}},
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	v := viewport.New(80, 8)
	v.SetContent("Ready to update...")

	return model{
		tasks:        tasks,
		currentTask:  0,
		spinner:      s,
		progress:     p,
		viewport:     v,
		outputChan:   make(chan string),
		finishedChan: make(chan error),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		startTask(m.tasks[0], m.outputChan, m.finishedChan),
		waitForActivity(m.outputChan, m.finishedChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 4
		m.viewport.Width = msg.Width - 4

	case spinner.TickMsg:
		var sCmd tea.Cmd
		m.spinner, sCmd = m.spinner.Update(msg)
		cmds = append(cmds, sCmd)

	case taskOutputMsg:
		// Append to viewport
		m.tasks[m.currentTask].Output += string(msg) + "\n"
		m.viewport.SetContent(m.tasks[m.currentTask].Output)
		m.viewport.GotoBottom()

		// Continue waiting for activity
		cmds = append(cmds, waitForActivity(m.outputChan, m.finishedChan))

	case taskFinishedMsg:
		// Mark current task done/error
		if msg != nil {
			m.tasks[m.currentTask].Status = StatusError
			m.tasks[m.currentTask].Error = msg
		} else {
			m.tasks[m.currentTask].Status = StatusDone
		}

		// Move to next task
		m.currentTask++
		prog := float64(m.currentTask) / float64(len(m.tasks))

		// Update progress bar
		cmds = append(cmds, m.progress.SetPercent(prog))

		if m.currentTask >= len(m.tasks) {
			m.done = true
			m.viewport.SetContent("All updates completed!\nPress 'q' to exit.")
			return m, tea.Batch(cmds...)
		}

		// Start next task
		cmds = append(cmds, startTask(m.tasks[m.currentTask], m.outputChan, m.finishedChan))
		cmds = append(cmds, waitForActivity(m.outputChan, m.finishedChan))

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.done {
		return docStyle.Render(fmt.Sprintf(
			"%s\n\nAll tasks finished.\n\n%s",
			titleStyle.Render(),
			m.progress.View(),
		))
	}

	// Task List
	s := ""
	for i, t := range m.tasks {
		cursor := " "
		if i == m.currentTask {
			cursor = m.spinner.View()
		}

		checked := " "
		if t.Status == StatusDone {
			checked = checkMark.String()
		} else if t.Status == StatusError {
			checked = crossMark.String()
		} else if i == m.currentTask {
			checked = "..." // pending
		}

		title := t.Name
		if i == m.currentTask {
			title = lipgloss.NewStyle().Bold(true).Render(t.Name)
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, checked, title)
	}

	return docStyle.Render(fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		titleStyle.Render(),
		s,
		m.progress.View(),
		lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(m.viewport.View()),
	))
}

// Commands / Subs

func startTask(t Task, outChan chan string, finChan chan error) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(t.Command, t.Args...)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			go func() { finChan <- err }()
			return nil
		}

		// Stream output
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			for scanner.Scan() {
				outChan <- scanner.Text()
			}
			go func() {
				err := cmd.Wait()
				finChan <- err
			}()
		}()

		return nil
	}
}

func waitForActivity(outChan chan string, finChan chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case line := <-outChan:
			return taskOutputMsg(line)
		case err := <-finChan:
			return taskFinishedMsg(err)
		}
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
