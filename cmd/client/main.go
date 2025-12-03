package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

const defaultBaseURL = "http://localhost:8080"

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type passwordRequest struct {
	ID       string  `json:"id"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Meta     *string `json:"meta,omitempty"`
}

type passwordResponse struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

type textRequest struct {
	ID   string  `json:"id"`
	Text string  `json:"text"`
	Meta *string `json:"meta,omitempty"`
}

type textResponse struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

type cardRequest struct {
	Number string  `json:"number"`
	Holder string  `json:"holder"`
	Expire string  `json:"expire"`
	CVV    string  `json:"cvv"`
	Meta   *string `json:"meta,omitempty"`
}

type cardResponse struct {
	CardPAN string  `json:"card_pan"`
	Number  string  `json:"number"`
	Holder  string  `json:"holder"`
	Expire  string  `json:"expire"`
	CVV     string  `json:"cvv"`
	Meta    *string `json:"meta,omitempty"`
}

type binaryRequest struct {
	ID   string  `json:"id"`
	Data string  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

type binaryResponse struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// общая функция для HTTP-клиента
func httpClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// doJSONRequest выполняет JSON-запрос с необязательным токеном и парсит ответ в out.
func doJSONRequest(method, url, token string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			return fmt.Errorf("encode payload: %w", err)
		}
		body = &buf
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// loginAndGetToken логинится и возвращает JWT-токен из заголовка Authorization.
func loginAndGetToken(baseURL, login, password string) (string, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/user/login"
	reqBody := loginRequest{
		Login:    login,
		Password: password,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return "", fmt.Errorf("encode login: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("do login request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("login failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	authHeader := resp.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no Authorization header in login response")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("unexpected Authorization header: %s", authHeader)
	}
	return strings.TrimPrefix(authHeader, prefix), nil
}

// -------- Команды --------

func cmdVersion() {
	fmt.Printf("GophKeeper client\nVersion: %s\nBuild date: %s\n", version, buildDate)
}

func cmdRegister(args []string) {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	login := fs.String("login", "", "User login")
	password := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *login == "" || *password == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/user/register"
	reqBody := loginRequest{Login: *login, Password: *password}
	var resp map[string]any

	if err := doJSONRequest(http.MethodPost, url, "", reqBody, &resp); err != nil {
		fmt.Println("register error:", err)
		os.Exit(1)
	}
	fmt.Println("Registered successfully:", resp)
}

func cmdLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	login := fs.String("login", "", "User login")
	password := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *login == "" || *password == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *login, *password)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}
	fmt.Println("Login OK. Token (JWT):")
	fmt.Println(token)
}

// --- Passwords ---

func cmdAddPassword(args []string) {
	fs := flag.NewFlagSet("add-password", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Password record ID")
	serviceLogin := fs.String("plogin", "", "Service login for this password")
	servicePassword := fs.String("ppass", "", "Service password")
	meta := fs.String("meta", "", "Optional meta information")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" || *serviceLogin == "" || *servicePassword == "" {
		fmt.Println("login, password, id, plogin and ppass are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	var metaPtr *string
	if *meta != "" {
		metaPtr = meta
	}

	reqBody := passwordRequest{
		ID:       *id,
		Login:    *serviceLogin,
		Password: *servicePassword,
		Meta:     metaPtr,
	}
	url := strings.TrimRight(*baseURL, "/") + "/api/v1/passwords"
	var resp passwordResponse

	if err := doJSONRequest(http.MethodPost, url, token, reqBody, &resp); err != nil {
		fmt.Println("add-password error:", err)
		os.Exit(1)
	}
	fmt.Printf("Password record saved: %+v\n", resp)
}

func cmdGetPassword(args []string) {
	fs := flag.NewFlagSet("get-password", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Password record ID")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" {
		fmt.Println("login, password and id are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/passwords/" + *id
	var resp passwordResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("get-password error:", err)
		os.Exit(1)
	}
	fmt.Printf("Password record: %+v\n", resp)
}

func cmdListPasswords(args []string) {
	fs := flag.NewFlagSet("list-passwords", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/passwords"
	var resp []passwordResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("list-passwords error:", err)
		os.Exit(1)
	}
	fmt.Printf("Password records: %+v\n", resp)
}

// --- Texts ---

func cmdAddText(args []string) {
	fs := flag.NewFlagSet("add-text", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Text ID (title)")
	text := fs.String("text", "", "Text content")
	meta := fs.String("meta", "", "Optional meta")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" || *text == "" {
		fmt.Println("login, password, id and text are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	var metaPtr *string
	if *meta != "" {
		metaPtr = meta
	}

	reqBody := textRequest{
		ID:   *id,
		Text: *text,
		Meta: metaPtr,
	}
	url := strings.TrimRight(*baseURL, "/") + "/api/v1/texts"
	var resp textResponse
	if err := doJSONRequest(http.MethodPost, url, token, reqBody, &resp); err != nil {
		fmt.Println("add-text error:", err)
		os.Exit(1)
	}
	fmt.Printf("Text record saved: %+v\n", resp)
}

func cmdGetText(args []string) {
	fs := flag.NewFlagSet("get-text", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Text ID")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" {
		fmt.Println("login, password and id are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/texts/" + *id
	var resp textResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("get-text error:", err)
		os.Exit(1)
	}
	fmt.Printf("Text record: %+v\n", resp)
}

func cmdListTexts(args []string) {
	fs := flag.NewFlagSet("list-texts", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/texts"
	var resp []textResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("list-texts error:", err)
		os.Exit(1)
	}
	fmt.Printf("Text records: %+v\n", resp)
}

// --- Cards ---

func cmdAddCard(args []string) {
	fs := flag.NewFlagSet("add-card", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	number := fs.String("number", "", "Card number")
	holder := fs.String("holder", "", "Card holder name")
	expire := fs.String("expire", "", "Expire date (e.g. 12/30)")
	cvv := fs.String("cvv", "", "CVV")
	meta := fs.String("meta", "", "Optional meta")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *number == "" || *holder == "" || *expire == "" || *cvv == "" {
		fmt.Println("login, password, number, holder, expire and cvv are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	var metaPtr *string
	if *meta != "" {
		metaPtr = meta
	}

	reqBody := cardRequest{
		Number: *number,
		Holder: *holder,
		Expire: *expire,
		CVV:    *cvv,
		Meta:   metaPtr,
	}
	url := strings.TrimRight(*baseURL, "/") + "/api/v1/cards"
	var resp cardResponse
	if err := doJSONRequest(http.MethodPost, url, token, reqBody, &resp); err != nil {
		fmt.Println("add-card error:", err)
		os.Exit(1)
	}
	fmt.Printf("Card saved: %+v\n", resp)
}

func cmdGetCard(args []string) {
	fs := flag.NewFlagSet("get-card", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	cardPAN := fs.String("card-pan", "", "Masked card PAN (e.g. 4600********5363)")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *cardPAN == "" {
		fmt.Println("login, password and card-pan are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/cards/" + *cardPAN
	var resp cardResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("get-card error:", err)
		os.Exit(1)
	}
	fmt.Printf("Card: %+v\n", resp)
}

func cmdListCards(args []string) {
	fs := flag.NewFlagSet("list-cards", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/cards"
	var resp []cardResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("list-cards error:", err)
		os.Exit(1)
	}
	fmt.Printf("Cards: %+v\n", resp)
}

// --- Binaries ---

func cmdAddBinary(args []string) {
	fs := flag.NewFlagSet("add-binary", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Binary record ID")
	data := fs.String("data", "", "Binary data as string (you can pre-encode base64)")
	meta := fs.String("meta", "", "Optional meta")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" || *data == "" {
		fmt.Println("login, password, id and data are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	var metaPtr *string
	if *meta != "" {
		metaPtr = meta
	}

	reqBody := binaryRequest{
		ID:   *id,
		Data: *data,
		Meta: metaPtr,
	}
	url := strings.TrimRight(*baseURL, "/") + "/api/v1/binaries"
	var resp binaryResponse
	if err := doJSONRequest(http.MethodPost, url, token, reqBody, &resp); err != nil {
		fmt.Println("add-binary error:", err)
		os.Exit(1)
	}
	fmt.Printf("Binary record saved: %+v\n", resp)
}

func cmdGetBinary(args []string) {
	fs := flag.NewFlagSet("get-binary", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	id := fs.String("id", "", "Binary record ID")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" || *id == "" {
		fmt.Println("login, password and id are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/binaries/" + *id
	var resp binaryResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("get-binary error:", err)
		os.Exit(1)
	}
	fmt.Printf("Binary record: %+v\n", resp)
}

func cmdListBinaries(args []string) {
	fs := flag.NewFlagSet("list-binaries", flag.ExitOnError)
	baseURL := fs.String("base-url", defaultBaseURL, "Base URL of server")
	userLogin := fs.String("login", "", "User login")
	userPassword := fs.String("password", "", "User password")
	_ = fs.Parse(args)

	if *userLogin == "" || *userPassword == "" {
		fmt.Println("login and password are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := loginAndGetToken(*baseURL, *userLogin, *userPassword)
	if err != nil {
		fmt.Println("login error:", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/api/v1/binaries"
	var resp []binaryResponse
	if err := doJSONRequest(http.MethodGet, url, token, nil, &resp); err != nil {
		fmt.Println("list-binaries error:", err)
		os.Exit(1)
	}
	fmt.Printf("Binary records: %+v\n", resp)
}

// -------- main / роутер команд --------

func printUsage() {
	fmt.Println("GophKeeper client")
	fmt.Println("Usage:")
	fmt.Println("  gophkeeper-client <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  version")
	fmt.Println("  register")
	fmt.Println("  login")
	fmt.Println("  add-password, get-password, list-passwords")
	fmt.Println("  add-text, get-text, list-texts")
	fmt.Println("  add-card, get-card, list-cards")
	fmt.Println("  add-binary, get-binary, list-binaries")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "version":
		cmdVersion()
	case "register":
		cmdRegister(args)
	case "login":
		cmdLogin(args)

	case "add-password":
		cmdAddPassword(args)
	case "get-password":
		cmdGetPassword(args)
	case "list-passwords":
		cmdListPasswords(args)

	case "add-text":
		cmdAddText(args)
	case "get-text":
		cmdGetText(args)
	case "list-texts":
		cmdListTexts(args)

	case "add-card":
		cmdAddCard(args)
	case "get-card":
		cmdGetCard(args)
	case "list-cards":
		cmdListCards(args)

	case "add-binary":
		cmdAddBinary(args)
	case "get-binary":
		cmdGetBinary(args)
	case "list-binaries":
		cmdListBinaries(args)

	default:
		fmt.Println("Unknown command:", cmd)
		printUsage()
		os.Exit(1)
	}
}
