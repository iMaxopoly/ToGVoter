package main

import (
	ui "github.com/gizak/termui"
	"runtime"
	"fmt"
	"strconv"
	"github.com/jimmysawczuk/worker"
	"sync"
)

var VotesAttempted int = 0
var ProxiesLeft int = 0
var SuccessfulVotes int = 0
var FailedVotes int = 0
var ErrorsEncountered int = 0
var LoadedProxies int = 0


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	l := &sync.Mutex{}
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	ui.UseTheme("helloworld")

	loadConfErr := LoadConfig()
	loadProxyErr:= LoadProxies()

	LoadedProxies = len(D_PROXY_LIST)
	ProxiesLeft = LoadedProxies
	aboutUi := ui.NewPar(
		"Please report bugs @skype: newjolt\nOptimized to use "+ strconv.Itoa(runtime.NumCPU()) +" cores.\nPress 'CTRL+C' to quit, 'CTRL+SPACE' to start task, 'CTRL+R' to reload settings.\nJoin us revive KryptoDEV.com")
	aboutUi.Height = 5
	aboutUi.TextFgColor = ui.ColorBlack
	aboutUi.Border.FgColor = ui.ColorMagenta
	aboutUi.TextBgColor = ui.ColorWhite
	aboutUi.BgColor = ui.ColorWhite
	aboutUi.Border.Label = "TOGNINJA: a topofgames.com auto voter v0.1 - kryptoDEV.com"
	aboutUi.Border.LabelFgColor = ui.AttrBold

	successUi := ui.NewPar("")
	successUi.Height = 3
	successUi.TextFgColor = ui.ColorWhite
	successUi.Border.Label = "Last Succeeded Vote"

	failedUi := ui.NewPar("")
	failedUi.Height = 3
	failedUi.TextFgColor = ui.ColorWhite
	failedUi.Border.Label = "Last Failed Vote"

	configUi := ui.NewPar(fmt.Sprintf("@Proxies Loaded: %d\n@Target ID: %d\n@Worker Count: %d", LoadedProxies, ConfigStruct.TargetID, ConfigStruct.WorkerCount))
	configUi.Height = 5
	configUi.TextFgColor = ui.ColorWhite
	configUi.Border.FgColor = ui.ColorWhite
	configUi.Border.Label = "Loaded Config."

	progressUi := ui.NewGauge()
	progressUi.Percent = 0
	progressUi.Height = 3
	progressUi.Border.Label = "Progress Report"
	progressUi.BarColor = ui.ColorRed
	progressUi.Border.FgColor = ui.ColorWhite
	progressUi.Border.LabelFgColor = ui.ColorCyan

	errorUi := ui.NewPar("")
	if loadConfErr != nil{
		errorUi.Text = loadConfErr.Error()
	} else if ConfigStruct.DBCUsername ==""{
		errorUi.Text = "Deathbycaptcha username not provided."
	} else if ConfigStruct.DBCPassword == ""{
		errorUi.Text = "Deathbycaptcha password not provided."
	} else if loadProxyErr != nil {
		errorUi.Text = loadProxyErr.Error()
	}
	errorUi.Height = 3
	errorUi.TextFgColor = ui.ColorRed
	errorUi.Border.FgColor = ui.ColorWhite
	errorUi.Border.Label = "Last Error"

	statsUi := ui.NewPar("")
	statsUi.Height = 3
	statsUi.TextFgColor = ui.ColorWhite
	statsUi.Border.FgColor = ui.ColorGreen
	statsUi.Border.Label = "Stats"

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, aboutUi)),
		ui.NewRow(
			ui.NewCol(6, 0, configUi), ui.NewCol(6, 0, errorUi)),
		ui.NewRow(
			ui.NewCol(6, 0, successUi), ui.NewCol(6, 0, failedUi)),
		ui.NewRow(
			ui.NewCol(12, 0, progressUi)),
		ui.NewRow(
			ui.NewCol(12, 0, statsUi)))

	// calculate layout
	ui.Body.Align()

	done := make(chan bool)
	redraw := make(chan bool)

	evt := ui.EventCh()
	ui.Render(ui.Body)
	worker.MaxJobs = ConfigStruct.WorkerCount
	w := worker.NewWorker()

	w.On(worker.JobFinished, func(args ...interface{}){
		l.Lock()
		ProxiesLeft--
		l.Unlock()
		pk := args[0].(*worker.Package)
		job := pk.Job().(*VoteJob)
		if job.HasError != ""{
			l.Lock()
			errorUi.Text = job.HasError
			ErrorsEncountered++
			l.Unlock()
		}
		if job.HasSuccessVoted{
			l.Lock()
			SuccessfulVotes++
			VotesAttempted++
			l.Unlock()
			if job.SuccessVoteText != ""{
				l.Lock()
				successUi.Text = job.SuccessVoteText
				l.Unlock()
			}
		}
		if job.HasFailVoted{
			l.Lock()
			FailedVotes++
			VotesAttempted++
			l.Unlock()
			if job.FailVoteText != ""{
				l.Lock()
				failedUi.Text = job.FailVoteText
				l.Unlock()
			}
		}
		if job.IsLast {
			l.Lock()
			statsUi.Text = "Bot: Finished | Proxies left: 0 | Successful Votes: "+strconv.Itoa(SuccessfulVotes)+" | Failed Votes: "+strconv.Itoa(FailedVotes)+" | Errors Encountered: "+strconv.Itoa(ErrorsEncountered)
			progressUi.Percent = 100
			progressUi.BarColor = ui.ColorGreen
			ui.Render(ui.Body)
			l.Unlock()
			return
		}
		l.Lock()
		progressUi.Percent = int(float32(100 - (100*ProxiesLeft/LoadedProxies)))
		statsUi.Text = "Bot: Running | Proxies left: "+strconv.Itoa(ProxiesLeft)+" | Successful Votes: "+strconv.Itoa(SuccessfulVotes)+" | Failed Votes: "+strconv.Itoa(FailedVotes)+" | Errors Encountered: "+strconv.Itoa(ErrorsEncountered)
		ui.Render(ui.Body)
		l.Unlock()
	})

	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey && e.Key == ui.KeyCtrlC {
				return
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyCtrlR {
				//FetchCaptcha()
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyCtrlSpace {
				for a := 0; a<LoadedProxies; a++ {
					if a == LoadedProxies {
						j := VoteJob{Name: D_PROXY_LIST[a], Proxy: D_PROXY_LIST[a], HasSuccessVoted: false, HasFailVoted: false, IsLast: true}
						w.Add(&j)
					} else {
						j := VoteJob{Name: D_PROXY_LIST[a], Proxy: D_PROXY_LIST[a], HasSuccessVoted: false, HasFailVoted: false, IsLast: false}
						w.Add(&j)
					}
				}
				w.RunUntilDone()
			}
			if e.Type == ui.EventResize {
				ui.Body.Width = ui.TermWidth()
				ui.Body.Align()
				go func() { redraw <- true }()
			}
		case <-done:
			return
		case <-redraw:
			ui.Render(ui.Body)
		}
	}
}