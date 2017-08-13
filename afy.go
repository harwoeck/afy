package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/op/go-logging"
	"golang.org/x/oauth2"
	apiGithub "golang.org/x/oauth2/github"
)

var log *logging.Logger
var config *configuration
var configGithub *oauth2.Config
var cookies *sessions.CookieStore
var users map[string]int
var page string
var tmpl *template.Template

func init() {
	log = logging.MustGetLogger("afy")
	format := logging.MustStringFormatter("%{color}[%{time:15:04:05.000}][%{shortfile}(%{shortfunc})-%{level:.4s}]%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFmt := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFmt)

	cnfg := flag.String("config", "", "Path to config file")
	flag.Parse()

	if cnfg == nil || *cnfg == "" {
		log.Error("Must provide parameter '-config filename.json'")
		os.Exit(1)
	}

	b, err := ioutil.ReadFile(*cnfg)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	cnt, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	page = string(cnt)
}

func main() {
	tmpl = template.Must(template.New("page").Parse(page))
	users = make(map[string]int)

	configGithub = &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		Scopes:       []string{},
		Endpoint:     apiGithub.Endpoint,
	}

	cookies = sessions.NewCookieStore(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
	cookies.Options = &sessions.Options{
		Path:     "/",
		Domain:   config.Host,
		MaxAge:   60 * 60 * 6,
		Secure:   true,
		HttpOnly: true,
	}

	http.HandleFunc("/", recoveryHandler(router))
	http.HandleFunc("/auth/github/login", recoveryHandler(githubLogin))
	http.HandleFunc("/auth/github/callback", recoveryHandler(githubCallback))
	http.HandleFunc("/f/", recoveryHandler(login))

	// Serve ether TLS or not
	if config.TLS {
		log.Fatal(http.ListenAndServeTLS(config.Addr, config.Cert, config.Key, nil))
	} else {
		log.Fatal(http.ListenAndServe(config.Addr, nil))
	}
}

func recoveryHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rval := recover(); rval != nil {
				debug.PrintStack()
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		handler(w, r)
	}
}
