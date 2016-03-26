package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var (
	email      = flag.String("email", "", "Service email")
	privateKey = flag.String("privateKey", "", "Private key")

	domain = flag.String("domain", "storage.googleapis.com", "Host defaults to `storage.googleapis.com`")
)

func main() {
	flag.Parse()

	if *email == "" {
		log.Fatalf("Email argument is required. See --help.")
	}
	if *privateKey == "" {
		log.Fatalf("PrivateKey argument is required. See --help.")
	}

	key, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		log.Fatalf("Unable to load %s: %v", *privateKey, err)
	}

	conf := &jwt.Config{
		Email:      *email,
		PrivateKey: []byte(key),
		Scopes:     []string{"https://www.googleapis.com/auth/devstorage.read_only"},
		TokenURL:   google.JWTTokenURL,
	}

	token, _ := conf.TokenSource(context.Background()).Token()

	var out io.Writer

	if flag.NArg() > 0 {
		filename := flag.Arg(0)
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, os.ModeDir); err != nil {
			log.Fatalf("Failed to create directory `%s`: %v", dir, err)
		}
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Failed to create file `%s`: %v", filename, err)
		}
		out = f
	} else {
		out = os.Stdout
	}

	enc := json.NewEncoder(out)

	type Credentials struct {
		Token string `json:"token"`
	}

	enc.Encode(struct {
		RktKind     string   `json:"rktKind"`
		RktVersion  string   `json:"rktVersion"`
		Domains     []string `json:"domains"`
		Type        string   `json:"type"`
		Credentials Credentials
	}{
		RktKind:     "auth",
		RktVersion:  "v1",
		Domains:     []string{*domain},
		Type:        "oauth",
		Credentials: Credentials{Token: token.AccessToken},
	})
}
