package updater

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/domain"
)

var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	checkMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	crossMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("✗")
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
)

type Model struct {
	tasks        []domain.Task
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
	Exit         bool
}

// Messages
type taskOutputMsg string
type taskFinishedMsg error

func NewModel(tasks []domain.Task) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	v := viewport.New(80, 8)
	v.SetContent("업데이트 준비 중...")

	return Model{
		tasks:        tasks,
		currentTask:  0,
		spinner:      s,
		progress:     p,
		viewport:     v,
		outputChan:   make(chan string),
		finishedChan: make(chan error),
	}
}

func (m Model) Init() tea.Cmd {
	if len(m.tasks) == 0 {
		m.done = true
		m.viewport.SetContent("업데이트할 항목이 없습니다.")
		return nil
	}
	return tea.Batch(
		m.spinner.Tick,
		startTask(m.tasks[0], m.outputChan, m.finishedChan),
		waitForActivity(m.outputChan, m.finishedChan),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			if m.done {
				m.Exit = true
				return m, nil
			}
			// While running, maybe ask to confirm or just quit app?
			// For now, let's allow quitting only when done or forced via ctrl+c (handled by app)
			if msg.String() == "q" && m.done {
				m.Exit = true
				return m, nil
			}
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
			m.tasks[m.currentTask].Status = domain.StatusError
			m.tasks[m.currentTask].Error = msg
		} else {
			m.tasks[m.currentTask].Status = domain.StatusDone
		}

		// Move to next task
		m.currentTask++
		prog := float64(m.currentTask) / float64(len(m.tasks))

		// Update progress bar
		cmds = append(cmds, m.progress.SetPercent(prog))

		if m.currentTask >= len(m.tasks) {
			m.done = true
			m.viewport.SetContent("모든 업데이트가 완료되었습니다!\n'q'를 눌러 나가세요.")
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

func (m Model) View() string {
	if m.done {
		// Task List Summary
		s := ""
		for _, t := range m.tasks {
			checked := " "
			if t.Status == domain.StatusDone {
				checked = checkMark.String()
			} else if t.Status == domain.StatusError {
				checked = crossMark.String()
			}
			s += fmt.Sprintf("%s %s\n", checked, t.Name)
		}

		return docStyle.Render(fmt.Sprintf(
			"업데이트 완료!\n\n%s\n\n%s\n\n(q: 돌아가기)",
			s,
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
		if t.Status == domain.StatusDone {
			checked = checkMark.String()
		} else if t.Status == domain.StatusError {
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
		"시스템 업데이트 진행 중...\n\n%s\n\n%s\n\n%s",
		s,
		m.progress.View(),
		lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(m.viewport.View()),
	))
}

// Commands / Subs

func startTask(t domain.Task, outChan chan string, finChan chan error) tea.Cmd {
	return func() tea.Msg {
		// For demo purposes, we might want to log the exact command being run
		outChan <- fmt.Sprintf("Running: %s %s", t.Command, strings.Join(t.Args, " "))

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
