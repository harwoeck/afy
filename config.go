package main

type yamlCnfg struct {
	Addr string
	Host string
	TLS  struct {
		Active bool
		Cert   string
		Key    string
	}
	Gzip bool
	Mnt  string
	ACP  struct {
		Github struct {
			ClientID     string
			ClientSecret string
		}
		Whitelist     bool
		WhitelistFile string
		Org           bool
		OrgName       string
	}
	Cookie struct {
		UseStatic bool
		AuthKey   string
		CryptKey  string
	}
}
