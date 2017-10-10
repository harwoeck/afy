package main

type ymlCnfg struct {
	Addr string
	Host string
	TLS  struct {
		Active bool
		Cert   string
		Key    string
	}
	Gzip    bool
	Mnt     string
	FSCache struct {
		RefreshTime int
	}
	ACP struct {
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
	Branding struct {
		Name        string
		Description string
		Keywords    string
		Favicon     string
	}
}
