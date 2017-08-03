package main

type configuration struct {
	Addr string `json:"addr"`

	Host string `json:"host"`
	TLS  bool   `json:"tls"`
	Cert string `json:"cert"`
	Key  string `json:"key"`

	Root string `json:"root"`

	GithubClientID     string `json:"github_client_id"`
	GithubClientSecret string `json:"github_client_secret"`
}
