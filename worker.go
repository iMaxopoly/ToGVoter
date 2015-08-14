package main
import (
	"time"
)

type VoteJob struct {
	Name                string
	Proxy               string
	IsLast              bool
	HasSuccessVoted    bool
	HasFailVoted       bool
	HasError           string
	SuccessVoteText   string
	FailVoteText      string
}

func (s *VoteJob) Run() {
	image, cookieReceived, err := FetchCaptcha(s.Proxy)
	if err != nil {
		s.HasError = err.Error()
		return
	}

	pollID, mlfao := SolveCaptcha1(image)
	if mlfao != nil {
		s.HasError = mlfao.Error()
		return
	}
	var solveResult string
	timer := time.NewTimer(20 * time.Second)
	var repoll bool
	repoll = true
	for {
		select {
		case <-timer.C:
			repoll = false
			break
		default:
			result, pollAgain, blmao := SolveCaptcha2(pollID)
			if blmao != nil {
				continue
			}
			if !pollAgain {
				solveResult = result
				repoll = false
				timer.Stop()
			}
		}
		if !repoll {
			break
		}
	}
	if solveResult != "" {
		if err := VoteTarget(s.Proxy, solveResult, cookieReceived); err != nil {
			if err.Error() == "Bad Captcha." {
				s.HasFailVoted = true
				s.FailVoteText = "Wrong captcha via "+s.Proxy
				return
			}
			s.HasError = err.Error()
			return
		}
		s.HasSuccessVoted = true
		s.SuccessVoteText = "Right captcha via "+s.Proxy
	}
}

