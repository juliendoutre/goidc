package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	version = "unknown"
	commit  = "unknown" //nolint:gochecknoglobals
)

func main() {
	port := flag.Uint64("port", 8000, "Port to listen on") //nolint:mnd
	decode := flag.Bool("decode", false, "Print out the decoded claims instead of the raw OIDC token")
	showVersion := flag.Bool("version", false, "Show this program's version and exit")
	provider := flag.String("provider", "https://gitlab.com", "OIDC provider")
	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "proxaudit %s (%s)\n", version, commit)

		return
	}

	os.Exit(run(*provider, *port, *decode))
}

func run(providerURL string, port uint64, decode bool) int {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	verifier := oauth2.GenerateVerifier()

	redirectURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", port),
		Path:   "/",
	}

	conf := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL.String(),
	}

	state := oauth2.GenerateVerifier()

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	if err := exec.CommandContext(ctx, "open", url).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	codeCh := make(chan string, 1)

	router := http.NewServeMux()
	router.HandleFunc("/", handle(codeCh, state))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	code := <-codeCh

	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	oauthToken, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	idToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		fmt.Fprintln(os.Stderr, "failed to read id_token")

		return 1
	}

	parsedIDToken, err := provider.Verifier(&oidc.Config{ClientID: os.Getenv("CLIENT_ID")}).Verify(ctx, idToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	if decode {
		var claims map[string]interface{}
		if err := parsedIDToken.Claims(&claims); err != nil {
			fmt.Fprintln(os.Stderr, err)

			return 1
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(claims); err != nil {
			fmt.Fprintln(os.Stderr, err)

			return 1
		}

		return 0
	}

	fmt.Fprint(os.Stdout, idToken)

	return 0
}

func handle(codeCh chan<- string, state string) func(w http.ResponseWriter, r *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		if query.Get("state") != state {
			res.WriteHeader(http.StatusForbidden)
			_, _ = res.Write([]byte("Incorrect state, request denied ❌"))
		}

		codeCh <- query.Get("code")

		res.WriteHeader(http.StatusFound)
		_, _ = res.Write([]byte("You can close this page now 😊"))

		close(codeCh)
	}
}
