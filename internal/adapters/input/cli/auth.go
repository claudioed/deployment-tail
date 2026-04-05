package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// NewAuthCmd creates the auth command with subcommands
func NewAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication management",
		Long:  "Manage authentication with the deployment-tail API",
	}

	authCmd.AddCommand(NewLoginCmd())
	authCmd.AddCommand(NewLogoutCmd())
	authCmd.AddCommand(NewAuthStatusCmd())

	return authCmd
}

// NewLoginCmd creates the login command
func NewLoginCmd() *cobra.Command {
	var manual bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to deployment-tail",
		Long:  "Authenticate with Google OAuth and store the JWT token locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(apiEndpoint, manual)
		},
	}

	cmd.Flags().BoolVar(&manual, "manual", false, "Manual authentication (for headless environments)")

	return cmd
}

// NewLogoutCmd creates the logout command
func NewLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from deployment-tail",
		Long:  "Revoke the current JWT token and remove local authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(apiEndpoint)
		},
	}

	return cmd
}

// NewAuthStatusCmd creates the auth status command
func NewAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Display current authentication status and user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus(apiEndpoint)
		},
	}

	return cmd
}

// runLogin performs the OAuth login flow
func runLogin(apiURL string, manual bool) error {
	// Start local callback server
	callbackChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Create callback server
	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			errChan <- fmt.Errorf("no authorization code in callback")
			return
		}

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
			<head><title>Authentication Successful</title></head>
			<body>
				<h1>✓ Authentication Successful</h1>
				<p>You can close this window and return to the terminal.</p>
			</body>
			</html>
		`)

		callbackChan <- code
	})

	// Start server in background
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Get OAuth URL from server
	authURL := fmt.Sprintf("%s/auth/google/login?redirect=http://localhost:8081/callback", apiURL)

	if manual {
		// Manual flow for headless environments
		fmt.Println("Manual authentication mode")
		fmt.Println("\nPlease open this URL in your browser:")
		fmt.Printf("\n  %s\n\n", authURL)
		fmt.Print("After authenticating, paste the authorization code here: ")

		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			return fmt.Errorf("failed to read code: %w", err)
		}

		return exchangeCodeForToken(apiURL, code)
	}

	// Automatic flow with browser
	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit: %s\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Println("Please open the URL manually")
	}

	// Wait for callback or error
	select {
	case code := <-callbackChan:
		return exchangeCodeForToken(apiURL, code)
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timeout")
	}
}

// exchangeCodeForToken exchanges the authorization code for a JWT token
func exchangeCodeForToken(apiURL, code string) error {
	// Call the API callback endpoint
	callbackURL := fmt.Sprintf("%s/auth/google/callback?code=%s&state=cli-auth", apiURL, url.QueryEscape(code))

	resp, err := http.Get(callbackURL)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s", string(body))
	}

	// Parse response
	var authResp struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Save token to local storage
	tokenStore := NewTokenStore()
	if err := tokenStore.SaveToken(authResp.Token, authResp.User.Email, authResp.User.Role, time.Now().Add(24*time.Hour)); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("\n✓ Authentication successful!")
	fmt.Printf("Logged in as: %s (%s)\n", authResp.User.Email, authResp.User.Role)
	fmt.Printf("Role: %s\n", authResp.User.Role)

	return nil
}

// runLogout revokes the token and removes local authentication
func runLogout(apiURL string) error {
	// Load token
	tokenStore := NewTokenStore()
	token, err := tokenStore.LoadToken()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Not currently authenticated")
			return nil
		}
		return fmt.Errorf("failed to load token: %w", err)
	}

	// Call logout endpoint
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/logout", apiURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Even if the API call fails, delete local token
		tokenStore.DeleteToken()
		return fmt.Errorf("failed to revoke token on server (token deleted locally): %w", err)
	}
	defer resp.Body.Close()

	// Delete local token
	if err := tokenStore.DeleteToken(); err != nil {
		return fmt.Errorf("failed to delete local token: %w", err)
	}

	fmt.Println("✓ Logged out successfully")
	return nil
}

// runAuthStatus displays current authentication status
func runAuthStatus(apiURL string) error {
	// Load token
	tokenStore := NewTokenStore()
	token, err := tokenStore.LoadToken()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Status: Not authenticated")
			fmt.Println("\nRun 'deployment-tail auth login' to authenticate")
			return nil
		}
		return fmt.Errorf("failed to load token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		fmt.Println("Status: Token expired")
		fmt.Printf("Expired: %s\n", token.ExpiresAt.Format(time.RFC3339))
		fmt.Println("\nRun 'deployment-tail auth login' to re-authenticate")
		return nil
	}

	// Fetch current user info from API
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/me", apiURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Status: Authenticated (cached info)\n")
		fmt.Printf("Email: %s\n", token.Email)
		fmt.Printf("Role: %s\n", token.Role)
		fmt.Printf("Token expires: %s\n", token.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("\nWarning: Unable to verify with server: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Println("Status: Token invalid or revoked")
		fmt.Println("\nRun 'deployment-tail auth login' to re-authenticate")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get user info: %s", string(body))
	}

	// Parse user info
	var user struct {
		Email       string    `json:"email"`
		Name        string    `json:"name"`
		Role        string    `json:"role"`
		LastLoginAt time.Time `json:"last_login_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to parse user info: %w", err)
	}

	fmt.Println("Status: ✓ Authenticated")
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Role: %s\n", user.Role)
	fmt.Printf("Token expires: %s\n", token.ExpiresAt.Format(time.RFC3339))

	return nil
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		// Try common Linux browsers
		for _, browser := range []string{"xdg-open", "sensible-browser", "gnome-open", "kde-open"} {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, url)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no browser found")
		}
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
