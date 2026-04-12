package bootstrapauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"flora-hive/lib"
)

// Run optionally creates a uServer-Auth system and registers a first admin via HTTP (same contract as
// [userver-filemgr bootstrap:auth](https://github.com/ferdn4ndo/userver-filemgr)).
//
// Env (see .env.example):
//   - USERVER_AUTH_HOST — required when any bootstrap action runs.
//   - USERVER_AUTH_SYSTEM_CREATION_TOKEN or SYSTEM_CREATION_TOKEN — POST /auth/system (Authorization: Token …).
//   - USERVER_AUTH_SYSTEM_NAME — system name (body for /auth/system and /auth/register).
//   - USERVER_AUTH_BOOTSTRAP_CUSTOM_SYSTEM_TOKEN — optional custom token in create-system JSON.
//   - USERVER_AUTH_SYSTEM_TOKEN — existing system API key (for register-only flows).
//   - USERVER_AUTH_USER / USERVER_AUTH_PASSWORD — POST /auth/register.
//   - USERVER_AUTH_BOOTSTRAP_IS_ADMIN — optional "0"/"false" for non-admin (default admin).
//
// Skip: SKIP_USERVER_AUTH_SETUP=1 or SKIP_AUTH_BOOTSTRAP=1.
func Run(w io.Writer, env lib.Env) error {
	if strings.TrimSpace(os.Getenv("SKIP_USERVER_AUTH_SETUP")) == "1" ||
		strings.TrimSpace(os.Getenv("SKIP_AUTH_BOOTSTRAP")) == "1" {
		_, _ = fmt.Fprintln(w, "SKIP_USERVER_AUTH_SETUP / SKIP_AUTH_BOOTSTRAP: skipping bootstrap:auth")
		return nil
	}

	createTok := strings.TrimSpace(os.Getenv("USERVER_AUTH_SYSTEM_CREATION_TOKEN"))
	if createTok == "" {
		createTok = strings.TrimSpace(os.Getenv("SYSTEM_CREATION_TOKEN"))
	}
	sysName := strings.TrimSpace(env.UserverAuthSystemName)
	customTok := strings.TrimSpace(os.Getenv("USERVER_AUTH_BOOTSTRAP_CUSTOM_SYSTEM_TOKEN"))
	systemAPIKey := strings.TrimSpace(env.UserverAuthSystemToken)
	user := strings.TrimSpace(os.Getenv("USERVER_AUTH_USER"))
	pass := os.Getenv("USERVER_AUTH_PASSWORD")

	if createTok == "" && sysName == "" && user == "" && pass == "" {
		_, _ = fmt.Fprintln(w, "bootstrap:auth: no USERVER_AUTH_SYSTEM_CREATION_TOKEN / USERVER_AUTH_* bootstrap vars set; nothing to do.")
		return nil
	}

	base := strings.TrimSpace(env.UserverAuthBase())
	if base == "" {
		return fmt.Errorf("bootstrap:auth: set USERVER_AUTH_HOST when using auth bootstrap env vars")
	}

	timeout := 60 * time.Second
	if s := strings.TrimSpace(os.Getenv("USERVER_AUTH_HTTP_TIMEOUT_SEC")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	client := &http.Client{Timeout: timeout}
	ctxBase := strings.TrimSuffix(base, "/")

	var err error
	systemAPIKey, err = phaseCreateSystem(w, client, ctxBase, createTok, sysName, customTok, systemAPIKey)
	if err != nil {
		return err
	}

	if user == "" && pass == "" {
		if createTok != "" {
			_, _ = fmt.Fprintln(w, "bootstrap:auth: no USERVER_AUTH_USER/PASSWORD — system step done.")
		}
		return nil
	}
	if user == "" || pass == "" {
		return fmt.Errorf("bootstrap:auth: set both USERVER_AUTH_USER and USERVER_AUTH_PASSWORD for register")
	}
	if sysName == "" {
		return fmt.Errorf("bootstrap:auth: USERVER_AUTH_SYSTEM_NAME is required for register")
	}
	if systemAPIKey == "" {
		return fmt.Errorf("bootstrap:auth: USERVER_AUTH_SYSTEM_TOKEN is required to register (or complete system create first)")
	}
	return phaseRegisterAdmin(w, client, ctxBase, sysName, user, pass, systemAPIKey)
}

func phaseCreateSystem(w io.Writer, client *http.Client, ctxBase, createTok, sysName, customTok string, priorSystemKey string) (systemAPIKey string, err error) {
	systemAPIKey = priorSystemKey
	if createTok == "" {
		return systemAPIKey, nil
	}
	if sysName == "" {
		return "", fmt.Errorf("bootstrap:auth: USERVER_AUTH_SYSTEM_NAME is required when USERVER_AUTH_SYSTEM_CREATION_TOKEN (or SYSTEM_CREATION_TOKEN) is set")
	}
	createBody := map[string]any{"name": sysName}
	if customTok != "" {
		createBody["token"] = customTok
	}
	rawBody, _ := json.Marshal(createBody)
	req, err := http.NewRequest(http.MethodPost, ctxBase+"/auth/system", bytes.NewReader(rawBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+createTok)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("bootstrap:auth POST /auth/system: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	switch resp.StatusCode {
	case http.StatusCreated:
		var out struct {
			Token string `json:"token"`
			Name  string `json:"name"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			return "", fmt.Errorf("bootstrap:auth: decode system response: %w", err)
		}
		if out.Token != "" {
			systemAPIKey = out.Token
			persistBootstrapEnv(w, map[string]string{
				"USERVER_AUTH_SYSTEM_TOKEN": systemAPIKey,
			}, true)
		}
		_, _ = fmt.Fprintf(w, "bootstrap:auth: created system %q (token persisted when enabled).\n", out.Name)
		return systemAPIKey, nil
	case http.StatusConflict:
		_, _ = fmt.Fprintf(w, "bootstrap:auth: system %q already exists.\n", sysName)
		if systemAPIKey == "" {
			_, _ = fmt.Fprintln(w, "bootstrap:auth: set USERVER_AUTH_SYSTEM_TOKEN if you need POST /auth/register next.")
		}
		return systemAPIKey, nil
	default:
		return "", fmt.Errorf("bootstrap:auth POST /auth/system: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func phaseRegisterAdmin(w io.Writer, client *http.Client, ctxBase, sysName, user, pass, systemAPIKey string) error {
	isAdmin := true
	if v := strings.TrimSpace(os.Getenv("USERVER_AUTH_BOOTSTRAP_IS_ADMIN")); v == "0" || strings.EqualFold(v, "false") {
		isAdmin = false
	}
	isAdminStr := "1"
	if !isAdmin {
		isAdminStr = "0"
	}
	regBody, _ := json.Marshal(map[string]any{
		"username":     user,
		"system_name":  sysName,
		"system_token": systemAPIKey,
		"password":     pass,
		"is_admin":     isAdmin,
	})
	req2, err := http.NewRequest(http.MethodPost, ctxBase+"/auth/register", bytes.NewReader(regBody))
	if err != nil {
		return err
	}
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	if err != nil {
		return fmt.Errorf("bootstrap:auth POST /auth/register: %w", err)
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(io.LimitReader(resp2.Body, 1<<20))
	switch resp2.StatusCode {
	case http.StatusCreated:
		_, _ = fmt.Fprintf(w, "bootstrap:auth: registered user %q on system %q (is_admin=%v).\n", user, sysName, isAdmin)
		persistAfterRegister(w, systemAPIKey, user, pass, isAdminStr)
		return nil
	case http.StatusConflict:
		_, _ = fmt.Fprintf(w, "bootstrap:auth: user %q already exists on system %q (OK).\n", user, sysName)
		persistAfterRegister(w, systemAPIKey, user, pass, isAdminStr)
		return nil
	default:
		return fmt.Errorf("bootstrap:auth POST /auth/register: status %d: %s", resp2.StatusCode, strings.TrimSpace(string(body2)))
	}
}

func persistAfterRegister(w io.Writer, systemAPIKey, user, pass, isAdminStr string) {
	persistBootstrapEnv(w, map[string]string{
		"USERVER_AUTH_SYSTEM_TOKEN":       systemAPIKey,
		"USERVER_AUTH_USER":               user,
		"USERVER_AUTH_PASSWORD":           pass,
		"USERVER_AUTH_BOOTSTRAP_IS_ADMIN": isAdminStr,
	}, false)
}
