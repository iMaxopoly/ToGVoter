package main
import (
	"fmt"
	"time"
	"bytes"
	"strings"
	"net/url"
	"net/http"
	"mime/multipart"
	"io"
	"github.com/parnurzeal/gorequest"
)

func VoteTarget(proxy string, captchaAnswer string, cookie *http.Cookie) error {
	request := gorequest.New().Timeout(ConfigStruct.Timeout*time.Second)
	_, body, errs := request.
	Post(Target).
	Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
	Set("Content-Type", "application/x-www-form-urlencoded").
	Set("Referer", Target).
	Set("Cookie", cookie.Name+"="+cookie.Value).
	Set("Origin", "http://topofgames.com").
	Set("User-Agent", RandUserAgent()).
	AddCookie(cookie).
	RedirectPolicy(func(r gorequest.Request, via []gorequest.Request) error {
		r.URL.Opaque = r.URL.Path
		return nil}).
	Send("vote=1&yes=Yes&captcha="+captchaAnswer).
	End()
	if len(errs) != 0 {
		var buffer bytes.Buffer
		for i := 0; i<len(errs); i++ {
			buffer.WriteString(errs[i].Error()+"\n")
		}
		return fmt.Errorf(buffer.String())
	}

	if strings.Contains(body, `class="red"`) {
		return fmt.Errorf("Bad Captcha.")
	}
	return nil
}

func SolveCaptcha2(pollID string) (string, bool, error) {
	request := gorequest.New().Timeout(ConfigStruct.Timeout*time.Second)
	_, body, errs := request.
	Get("http://api.dbcapi.me/api/captcha/"+pollID).
	RedirectPolicy(func(r gorequest.Request, via []gorequest.Request) error {
		r.URL.Opaque = r.URL.Path
		return nil}).
	End()
	if len(errs) != 0 {
		var buffer bytes.Buffer
		for i := 0; i<len(errs); i++ {
			buffer.WriteString(errs[i].Error()+"\n")
		}
		return "", false, fmt.Errorf(buffer.String())
	}

	response, resperr := url.ParseQuery(body)
	if resperr != nil {
		return "", false, resperr
	}

	var cSol string

	if val, ok := response["text"]; !ok {
		return "", false, fmt.Errorf("Weird response from DBC2")
	} else {
		cSol = strings.TrimSpace((string)(val[0]))
	}
	if cSol == "" {
		return "", true, nil
	}
	return cSol, false, nil
}

func SolveCaptcha1(image string) (string, error) {
	// Create buffer
	buf := new(bytes.Buffer) // caveat IMO dont use this for large files, \
	// create a tmpfile and assemble your multipart from there (not tested)
	w := multipart.NewWriter(buf)
	// Create a form field writer for field label
	username, err := w.CreateFormField("username")
	if err != nil {
		return "", err
	}
	// Write label field
	username.Write([]byte(ConfigStruct.DBCUsername))
	// Create a form field writer for field summary
	password, err := w.CreateFormField("password")
	if err != nil {
		return "", err
	}
	// Write summary field
	password.Write([]byte(ConfigStruct.DBCPassword))
	// Create file field
	fw, err := w.CreateFormFile("captchafile", "image.jpg")
	if err != nil {
		return "", err
	}

	captcha := bytes.NewBufferString(image)
	_, err = io.Copy(fw, captcha)
	if err != nil {
		return "", err
	}
	// Important if you do not close the multipart writer you will not have a
	// terminating boundry
	w.Close()
	req, err := http.NewRequest("POST", "http://api.dbcapi.me/api/captcha", buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{
		Timeout:time.Duration(ConfigStruct.Timeout * time.Second),
	}
	resp, herr := client.Do(req)
	if herr != nil {
		return "", herr
	}
	body := &bytes.Buffer{}
	_, berr := body.ReadFrom(resp.Body)
	if berr != nil {
		return "", berr
	}
	resp.Body.Close()
	response, resperr := url.ParseQuery(body.String())
	if resperr != nil {
		return "", resperr
	}
	//fmt.Println(response["captcha"][0])
	var cSol string

	if val, ok := response["captcha"]; !ok {
		if eCreds, eOk := response["error"]; eOk {
			if strings.TrimSpace((string)(eCreds[0])) == "insufficient-funds" {
				return "", fmt.Errorf("DBC: Insufficient Funds")
			}
			if strings.TrimSpace((string)(eCreds[0])) == "not-logged-in" {
				return "", fmt.Errorf("DBC: Wrong user/pass")
			}
			if strings.TrimSpace((string)(eCreds[0])) == "invalid-captcha" {
				return "", fmt.Errorf("DBC: Invalid Captcha Provided")
			}
		}
		fmt.Println(response)
		return "", fmt.Errorf("Weird response from DBC1")
	} else {
		cSol = strings.TrimSpace((string)(val[0]))
	}
	return cSol, nil
}

func FetchCaptcha(proxy string) (string, *http.Cookie, error) {
	request := gorequest.New().Timeout(ConfigStruct.Timeout*time.Second)
	_, body, errs := request.
	Get(D_CAPTCHA_URL).
	Set("Referer", Target).
	Set("Accept", "image/webp,*/*;q=0.8").
	Set("Accept-Language", "en-US,en;q=0.8").
	Set("Connection", "keep-alive").
	Set("DNT", "1").
	Set("Host", "topofgames.com").
	Set("User-Agent", RandUserAgent()).
	RedirectPolicy(func(r gorequest.Request, via []gorequest.Request) error {
		r.URL.Opaque = r.URL.Path
		return nil}).
	End()
	if len(errs) != 0 {
		var buffer bytes.Buffer
		for i := 0; i<len(errs); i++ {
			buffer.WriteString(errs[i].Error()+"\n")
		}
		return "", nil, fmt.Errorf(buffer.String())
	}

	domain, _ := url.Parse(Target)
	if len(request.Client.Jar.Cookies(domain)) == 0 {
		return "", nil, fmt.Errorf("No Cookies Received.")
	}

	cookie := request.Client.Jar.Cookies(domain)[0]
	return body, cookie, nil
}