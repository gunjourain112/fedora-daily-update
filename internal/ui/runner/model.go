package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/domain"
	"my-updater/internal/ui/common"
)

type Model struct {
	Task     domain.Task
	Spinner  spinner.Model
	Viewport viewport.Model

	Output       string
	Done         bool
	Err          error

	outputChan   chan string
	finishedChan chan error

	Exit         bool
}

// Messages
type taskOutputMsg string
type taskFinishedMsg error

func NewModel(t domain.Task) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	v := viewport.New(80, 20)
	v.SetContent(fmt.Sprintf("Running %s...", t.Name))

	return Model{
		Task:         t,
		Spinner:      s,
		Viewport:     v,
		outputChan:   make(chan string),
		finishedChan: make(chan error),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		startTask(m.Task, m.outputChan, m.finishedChan),
		waitForActivity(m.outputChan, m.finishedChan),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Done {
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" {
				m.Exit = true
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width - 4
		m.Viewport.Height = msg.Height - 8 // Reserve space for header/footer

	case spinner.TickMsg:
		if !m.Done {
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case taskOutputMsg:
		m.Output += string(msg) + "\n"
		m.Viewport.SetContent(m.Output)
		m.Viewport.GotoBottom()
		cmds = append(cmds, waitForActivity(m.outputChan, m.finishedChan))

	case taskFinishedMsg:
		m.Done = true
		if msg != nil {
			m.Err = msg
			m.Output += fmt.Sprintf("\n\n❌ Error: %v\n", msg)
		} else {
			m.Output += "\n\n✅ 완료됨 (Completed)\n"
		}
		m.Viewport.SetContent(m.Output)
		m.Viewport.GotoBottom()
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	header := ""
	if m.Done {
		if m.Err != nil {
			header = common.TitleStyle.Foreground(common.ColorError).Render("오류 발생")
		} else {
			header = common.TitleStyle.Foreground(common.ColorSuccess).Render("작업 완료")
		}
	} else {
		header = fmt.Sprintf("%s %s", m.Spinner.View(), common.TitleStyle.Render(m.Task.Name))
	}

	footer := "\n(Enter: 목록으로 돌아가기)"
	if !m.Done {
		footer = ""
	}

	return common.DocStyle.Render(fmt.Sprintf(
		"%s\n\n%s%s",
		header,
		lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(m.Viewport.View()),
		footer,
	))
}

// Helpers

func startTask(t domain.Task, outChan chan string, finChan chan error) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(t.Command, t.Args...)

		// Capture stdout and stderr
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			go func() { finChan <- err }()
			return nil
		}

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
