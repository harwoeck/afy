package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gorilla/sessions"
	logging "github.com/op/go-logging"
	"golang.org/x/oauth2"
	apiGithub "golang.org/x/oauth2/github"
)

var fs http.Handler
var log *logging.Logger
var config *configuration
var configGithub *oauth2.Config
var cookies *sessions.CookieStore
var users map[string]int
var page string

const (
	cookieOAuthState = "_oast"
)

type configuration struct {
	Addr               string `json:"addr"`
	Host               string `json:"host"`
	Cert               string `json:"cert"`
	Key                string `json:"key"`
	Root               string `json:"root"`
	SessionAuth        string `json:"session_auth"`
	SessionCryp        string `json:"session_cryp"`
	SessionDomain      string `json:"session_domain"`
	GithubClientID     string `json:"github_client_id"`
	GithubClientSecret string `json:"github_client_secret"`
}

func init() {
	log = logging.MustGetLogger("vbgs")
	format := logging.MustStringFormatter("%{color}[%{time:15:04:05.000}][%{shortfile}(%{shortfunc})-%{level:.4s}]%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFmt := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFmt)

	cnfg := flag.String("config", "config.json", "File system location pointing to the gameserver instance configuration")
	flag.Parse()

	b, err := ioutil.ReadFile(*cnfg)
	if err != nil {
		log.Error(err.Error())
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Error(err.Error())
	}

	cnt, _ := ioutil.ReadFile("index.html")
	page = string(cnt)
}

func main() {
	tmpl := template.Must(template.New("page").Parse(page))
	http.HandleFunc("/fs/", func(w http.ResponseWriter, r *http.Request) {
		res, isDwld, isFrbn, is404, is500 := getDirList("/fs/", strings.TrimPrefix(r.URL.Path, "/fs/"))
		if isFrbn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if is404 {
			log.Warning("not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if is500 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if isDwld {
			log.Critical("ISDWLD")
			log.Info(r.URL.Path)
			http.StripPrefix("/fs/", fs).ServeHTTP(w, r)
		}
		tmpl.Execute(w, res)
	})
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Error(err.Error())
	}
	return

	users = make(map[string]int)
	fs = http.FileServer(http.Dir(config.Root))

	configGithub = &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  config.Host + "/",
		Scopes: []string{
			"user:email",
		},
		Endpoint: apiGithub.Endpoint,
	}

	authByte, err := base64.StdEncoding.DecodeString(config.SessionAuth)
	if err != nil {
		log.Error(err.Error())
		return
	}
	crypByte, err := base64.StdEncoding.DecodeString(config.SessionCryp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	cookies = sessions.NewCookieStore(authByte, crypByte)
	cookies.Options = &sessions.Options{
		Path:     "/",
		Domain:   config.SessionDomain,
		MaxAge:   60 * 60 * 24 * 7 * 4 * 12,
		Secure:   true,
		HttpOnly: true,
	}

	http.HandleFunc("/", recoveryHandler(githubLogin))
	http.HandleFunc("/auth/github/callback", recoveryHandler(githubCallback))
	http.HandleFunc("/fs/", recoveryHandler(assertLogin))

	if strings.TrimSpace(config.Cert) != "" {
		err = http.ListenAndServeTLS(config.Addr, config.Cert, config.Key, nil)
	} else {
		err = http.ListenAndServe(config.Addr, nil)
	}
	if err != nil {
		log.Error(err.Error())
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

func assertLogin(w http.ResponseWriter, r *http.Request) {
	/*url := r.URL.Path[len("/fs/"):]
	if !strings.Contains(url, "/") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if strings.Index(url, "/") != 32 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var id int
	var ok bool
	if id, ok = users[url[:32]]; !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	log.Infof("access: %d -> '%s'", id, url[32:])*/
	fs.ServeHTTP(w, r)
}
