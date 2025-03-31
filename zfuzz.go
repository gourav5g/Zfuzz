package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Zfuzz Banner with better visual design
const banner = `
███████╗████████╗██╗   ██╗███████╗███████╗████████╗███████╗██████╗ ██╗
██╔════╝╚══██╔══╝██║   ██║██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗██║
███████╗   ██║   ██║   ██║███████╗███████╗   ██║   █████╗  ██████╔╝██║
╚════██║   ██║   ██║   ██║╚════██║╚════██║   ██║   ██╔══╝  ██╔══██╗██║
███████║   ██║   ╚██████╔╝███████║███████║   ██║   ███████╗██║  ██║██████╗
╚══════╝   ╚═╝    ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═════╝
                                Made by Gourav
`

// Zfuzz options struct
type ZfuzzOptions struct {
	TargetURL   string
	Wordlist    string
	Threads     int
	ShowDetails bool
	Method      string
	Timeout     int
	Output      string
	OTP         bool
	APITest     bool
}

// Fuzz result structure for reporting
type FuzzResult struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
}

// OTP Bypass simulation
func otpBypass(url, otpPattern string, details bool) {
	// Simulate OTP bypass attempts by testing known OTP patterns
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < 1000; i++ {
		otp := fmt.Sprintf(otpPattern, i)
		fullURL := strings.Replace(url, "{OTP}", otp, 1)

		req, err := http.NewRequest("POST", fullURL, nil)
		if err != nil {
			if details {
				color.Red("[ERROR] %s\n", err)
			}
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			if details {
				color.Red("[ERROR] %s\n", err)
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			color.Green("[SUCCESS] OTP bypass succeeded with OTP: %s\n", otp)
			return
		} else {
			if details {
				color.Yellow("[FAILED] OTP attempt %s : %s\n", otp, resp.Status)
			}
		}
	}
	color.Red("[ERROR] OTP bypass failed after 1000 attempts.\n")
}

// API Penetration testing for authentication and other checks
func apiPenTest(url, method, token string, timeout int, details bool, results chan<- FuzzResult) {
	// Example API penetration testing
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	var req *http.Request
	var err error

	// Assuming token-based authentication
	if method == "POST" || method == "PUT" {
		req, err = http.NewRequest(method, url, nil)
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req, err = http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+token)
	}

	if err != nil {
		if details {
			color.Red("[ERROR] %s\n", err)
		}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		if details {
			color.Red("[ERROR] %s\n", err)
		}
		return
	}
	defer resp.Body.Close()

	result := FuzzResult{URL: url, Status: resp.StatusCode}
	results <- result

	// Output results
	if resp.StatusCode == 200 {
		color.Green("[FOUND] API Auth Success: %s\n", url)
	} else {
		color.Magenta("[FAILED] API Auth Failed: %s\n", url)
	}
}

// Read wordlist from file
func readWordlist(file string) ([]string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	return lines, nil
}

// Save results to CSV file
func saveToCSV(results []FuzzResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	writer.Write([]string{"URL", "Status"})

	// Write each result
	for _, result := range results {
		writer.Write([]string{result.URL, fmt.Sprintf("%d", result.Status)})
	}
	return nil
}

// Save results to JSON file
func saveToJSON(results []FuzzResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func main() {
	// Print the banner with color
	color.Cyan(banner)

	// Parse flags for input parameters
	var targetURL, wordlist, method, output, otpPattern, token string
	var threads, timeout int
	var showDetails, otpBypass, apiTest bool
	flag.StringVar(&targetURL, "url", "", "Target URL (e.g. http://example.com/{FUZZ})")
	flag.StringVar(&wordlist, "w", "", "Wordlist file path")
	flag.StringVar(&method, "m", "GET", "HTTP method (GET, POST, PUT)")
	flag.IntVar(&threads, "t", 10, "Number of threads (default: 10)")
	flag.IntVar(&timeout, "timeout", 10, "Timeout for each request (in seconds)")
	flag.BoolVar(&showDetails, "v", false, "Show detailed responses")
	flag.StringVar(&output, "o", "", "Output file (CSV/JSON)")
	flag.BoolVar(&otpBypass, "otp", false, "Enable OTP Bypass testing")
	flag.BoolVar(&apiTest, "api", false, "Enable API Penetration testing")
	flag.StringVar(&otpPattern, "otp-pattern", "123456", "OTP pattern to attempt")
	flag.StringVar(&token, "token", "", "API token for penetration testing")
	flag.Parse()

	if targetURL == "" || wordlist == "" {
		color.Red("[ERROR] Target URL and Wordlist are required.")
		os.Exit(1)
	}

	// Read wordlist
	words, err := readWordlist(wordlist)
	if err != nil {
		color.Red("[ERROR] Failed to read wordlist: %s\n", err)
		os.Exit(1)
	}

	// Create channel to store results
	results := make(chan FuzzResult)

	// Start OTP Bypass if enabled
	if otpBypass {
		otpBypass(targetURL, otpPattern, showDetails)
		return
	}

	// Start API Penetration testing if enabled
	if apiTest {
		go func() {
			for _, word := range words {
				apiPenTest(targetURL, method, token, timeout, showDetails, results)
			}
		}()
	}

	// Prepare for concurrent fuzzing
	var wg sync.WaitGroup
	wg.Add(len(words) * threads)

	// Start fuzzing using goroutines
	for i := 0; i < threads; i++ {
		go func() {
			for _, word := range words {
				fullURL := strings.Replace(targetURL, "{FUZZ}", word, 1)

				// Perform fuzzing for each URL
				client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
				req, err := http.NewRequest(method, fullURL, nil)
				if err != nil {
					if showDetails {
						color.Red("[ERROR] %s\n", err)
					}
					return
				}

				resp, err := client.Do(req)
				if err != nil {
					if showDetails {
						color.Red("[ERROR] %s\n", err)
					}
					return
				}
				defer resp.Body.Close()

				result := FuzzResult{URL: fullURL, Status: resp.StatusCode}
				results <- result

				// Output results
				if resp.StatusCode == 200 {
					color.Green("[FOUND] %s : %s\n", fullURL, resp.Status)
				} else if resp.StatusCode == 404 {
					color.Yellow("[NOT FOUND] %s : %s\n", fullURL, resp.Status)
				} else {
					color.Magenta("[OTHER STATUS] %s : %s\n", fullURL, resp.Status)
				}
			}
		}()
	}

	// Collect results
	var resultList []FuzzResult
	go func() {
		for result := range results {
			resultList = append(resultList, result)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// Save results to file (CSV/JSON)
	if output != "" {
		if strings.HasSuffix(output, ".csv") {
			err := saveToCSV(resultList, output)
			if err != nil {
				color.Red("[ERROR] Failed to save results to CSV: %s\n", err)
			} else {
				color.Green("[INFO] Results saved to CSV: %s\n", output)
			}
		} else if strings.HasSuffix(output, ".json") {
			err := saveToJSON(resultList, output)
			if err != nil {
				color.Red("[ERROR] Failed to save results to JSON: %s\n", err)
			} else {
				color.Green("[INFO] Results saved to JSON: %s\n", output)
			}
		} else {
			color.Red("[ERROR] Unsupported output format. Please use CSV or JSON.")
		}
	}
}
