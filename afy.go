package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/op/go-logging"
	"golang.org/x/oauth2"
	apiGithub "golang.org/x/oauth2/github"
	"gopkg.in/yaml.v2"
)

// Logger instance
var log *logging.Logger

// Configs
var config *yamlCnfg
var configGithub *oauth2.Config

// Cookie storage
var cookies *sessions.CookieStore

// Login user storage
var keyUser map[string]int
var userKey map[int]string

// ACP whitelist
var whitelist []int

// Template for index page
var tmpl *template.Template

func init() {
	// Initialize custom logger
	log = logging.MustGetLogger("afy")
	format := logging.MustStringFormatter("%{time:15:04:05.000} %{color}[%{level:.4s}][%{longfunc}]%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFmt := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFmt)

	// Define and parse flags
	flagConfig := flag.String("config", "", "Path to config file")
	flagGenkey := flag.Bool("genkey", false, "Used to generate random keys for authentication and encryption of cookies. Should be set in the config's cookie seetings")
	flagIndex := flag.String("indexpage", "index.tmpl", "")
	flag.Parse()

	// If afy was started with "-genkey" the user wants some secure random keys
	if *flagGenkey {
		log.Infof("Auth-Key:  %s", base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)))
		log.Infof("Crypt-Key: %s", base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)))
		os.Exit(0)
	}

	// Check if required flag "-config" was passed and load config
	if flagConfig == nil || *flagConfig == "" {
		log.Fatalf("Required parameter '-config yourConfigFileName.yml' is missing")
	}
	assertFSFilePtr(flagConfig)
	content, err := ioutil.ReadFile(*flagConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	config = &yamlCnfg{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Convert the mnt var from the yml file to an absolute path and check
	// if the directory exists
	assertFSDirPtr(&config.Mnt)

	// Convert the cert and key file paths to absolute paths and check if
	// the files exist
	if config.TLS.Active {
		assertFSFilePtr(&config.TLS.Cert)
		assertFSFilePtr(&config.TLS.Key)
	}

	// Read the template file used for all pages
	assertFSFilePtr(flagIndex)
	content, err = ioutil.ReadFile(*flagIndex)
	if err != nil {
		log.Fatal(err.Error())
	}
	tmpl, err = template.New("index").Parse(string(content))
	if err != nil {
		log.Fatal(err.Error())
	}

	// Initialize out store for user authentications
	keyUser = make(map[string]int)
	userKey = make(map[int]string)

	// Initialize our cookie store with strong authentication and encryption keys.
	if config.Cookie.UseStatic {
		cookies = sessions.NewCookieStore(base64Must(config.Cookie.AuthKey), base64Must(config.Cookie.CryptKey))
	} else {
		cookies = sessions.NewCookieStore(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32))
	}
	cookies.Options = &sessions.Options{
		Path:     "/",
		Domain:   config.Host,
		MaxAge:   60,
		Secure:   true,
		HttpOnly: true,
	}

	// Initialize our github oauth config
	scopes := []string{}
	if config.ACP.Org {
		scopes = append(scopes, "read:org")
	}
	configGithub = &oauth2.Config{
		ClientID:     config.ACP.Github.ClientID,
		ClientSecret: config.ACP.Github.ClientSecret,
		Scopes:       scopes,
		Endpoint:     apiGithub.Endpoint,
	}

	// Load favicon from base64 src
	favicon = base64Must(faviconSrc)

	// Load whitelist if enabled
	if config.ACP.Whitelist {
		assertFSDirPtr(&config.ACP.WhitelistFile)
		f, err := os.Open(config.ACP.WhitelistFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			cur := scanner.Text()
			if strings.Contains(cur, "#") {
				cur = strings.TrimSpace(strings.Split(cur, "#")[0])
			}
			curID, err := strconv.Atoi(cur)
			if err != nil {
				log.Error(err.Error())
			}
			whitelist = append(whitelist, curID)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func main() {
	// Register all handles defined below
	handles := map[string]func(http.ResponseWriter, *http.Request){
		"/":                     router,
		"/auth/github/login":    githubLogin,
		"/auth/github/callback": githubCallback,
		"/f/": authLayer,
	}
	for pattern, patternFunc := range handles {
		if config.Gzip {
			http.Handle(pattern, gziphandler.GzipHandler(http.HandlerFunc(recoveryHandler(patternFunc))))
		} else {
			http.HandleFunc(pattern, recoveryHandler(patternFunc))
		}
	}

	// Serve ether TLS or not
	if config.TLS.Active {
		log.Fatal(http.ListenAndServeTLS(config.Addr, config.TLS.Cert, config.TLS.Key, nil))
	} else {
		log.Fatal(http.ListenAndServe(config.Addr, nil))
	}
}
