package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

const (
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

type SearchPattern struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
}

func main() {
	urlFlag := flag.String("u", "", "A single URL to make the request")
	fileFlag := flag.String("f", "", "A file containing a list of URLs")
	colorFlag := flag.Bool("c", false, "Enable colored output")
	verboseFlag := flag.Bool("v", false, "display everything in verbose mode")
	flag.Parse()

	var body string
	uniqueValues := make(map[string]struct{})
	var baseURL string
	outputFilePath := "js_endpoint"
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error creating the output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	if *urlFlag != "" {
		baseURL, _ = extractBaseURL(*urlFlag)
		if isJavaScriptURL(*urlFlag) {
			responseBody, err := makeHTTPRequest(*urlFlag)
			if err != nil {
				fmt.Printf("Error making the HTTP request for %s: %v\n", *urlFlag, err)
			} else {
				fmt.Printf("[%sDone%s] %s\n", colorGreen, colorReset, *urlFlag)
				body = responseBody
				if err := run(body, *colorFlag, *verboseFlag, *urlFlag); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				lines := strings.Split(responseBody, "\n")
				for _, line := range lines {
					processDoubleQuotedStrings(line, baseURL, uniqueValues, outputFile, *verboseFlag, *colorFlag)
				}
			}
		}
	}

	if *fileFlag != "" {
		file, err := os.Open(*fileFlag)
		if err != nil {
			fmt.Printf("Error opening the file %s: %v\n", *fileFlag, err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Text()
			if isJavaScriptURL(url) {
				baseURL, _ = extractBaseURL(url)
				responseBody, err := makeHTTPRequest(url)
				if err != nil {
					fmt.Printf("Error making the HTTP request for %s: %v\n", url, err)
				} else {
					fmt.Printf("[%sDone%s] %s\n", colorGreen, colorReset, url)
					lines := strings.Split(responseBody, "\n")
					body = responseBody
					if err := run(body, *colorFlag, *verboseFlag, *urlFlag); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					for _, line := range lines {
						processDoubleQuotedStrings(line, baseURL, uniqueValues, outputFile, *verboseFlag, *colorFlag)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading from the file %s: %v\n", *fileFlag, err)
		}
	}

	if *urlFlag == "" && *fileFlag == "" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			url := scanner.Text()
			if isJavaScriptURL(url) {
				baseURL, _ = extractBaseURL(url)
				responseBody, err := makeHTTPRequest(url)
				if err != nil {
					fmt.Printf("Error making the HTTP request for %s: %v\n", url, err)
				} else {
					fmt.Printf("[%sDone%s] %s\n", colorGreen, colorReset, url)
					body = responseBody
					if err := run(body, *colorFlag, *verboseFlag, url); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					lines := strings.Split(responseBody, "\n")
					for _, line := range lines {
						processDoubleQuotedStrings(line, baseURL, uniqueValues, outputFile, *verboseFlag, *colorFlag)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from stdin:", err)
		}
	}
}

func isJavaScriptURL(url string) bool {
	return strings.HasSuffix(url, ".js")
}

func makeHTTPRequest(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(body), nil
	}

	return "", nil
}

func extractBaseURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Scheme + "://" + parsedURL.Host, nil
}

func run(body string, colorFlag, verboseFlag bool, urlFlag string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	jsonFilePath := filepath.Join(homeDir, ".config", "secret.json")

	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		return fmt.Errorf("Missing JSON file in the folder directory: %s", jsonFilePath)
	}

	patterns, err := readPatternFile(jsonFilePath)
	if err != nil {
		return err
	}

	matches := findMatches(body, patterns)

	if verboseFlag {
		displayMatches(matches, colorFlag)
	}

	if err := saveResultsToFile(matches, urlFlag); err != nil {
		return err
	}

	return nil
}

func readPatternFile(filePath string) ([]SearchPattern, error) {
	jsonContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var patterns []SearchPattern
	if err := json.Unmarshal(jsonContent, &patterns); err != nil {
		return nil, err
	}

	return patterns, nil
}

func findMatches(body string, patterns []SearchPattern) map[string][]string {
	matches := make(map[string][]string)

	for _, pattern := range patterns {
		for _, p := range pattern.Patterns {
			re := regexp.MustCompile(p)
			foundMatches := re.FindAllString(body, -1)
			if len(foundMatches) > 0 {
				matches[pattern.Name] = append(matches[pattern.Name], foundMatches...)
			}
		}
	}

	return matches
}

func displayMatches(matches map[string][]string, colorFlag bool) {
	for name, foundMatches := range matches {
		fmt.Printf("Matches for pattern (%s):\n", name)
		for _, match := range foundMatches {
			if colorFlag {
				red := color.New(color.FgRed).SprintFunc()
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Printf("%s ::: %s\n", red(name), green(match))
			} else {
				fmt.Printf("%s ::: %s\n", name, match)
			}
		}
	}
}

func saveResultsToFile(matches map[string][]string, urlFlag string) error {
	filePath := "js_secret"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	burl := urlFlag + "\n"

	_, err = file.WriteString(burl)
	if err != nil {
		return err
	}

	for name, foundMatches := range matches {
		for _, match := range foundMatches {
			line := name + " ::: " + match + "\n"
			_, err = file.WriteString(line)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processDoubleQuotedStrings(input string, baseURL string, uniqueValues map[string]struct{}, outputFile *os.File, verboseFlag bool, colorFlag bool) {
	regexPattern := `"/[a-zA-Z0-9_?&=\/\-#]*"`
	re := regexp.MustCompile(regexPattern)

	secondRegexPattern := `^"(.*)"$`
	secondRe := regexp.MustCompile(secondRegexPattern)

	matches := re.FindAllString(input, -1)
	for _, match := range matches {
		submatches := secondRe.FindStringSubmatch(match)
		if len(submatches) > 1 {
			endpoint := submatches[1]
			completeURL := baseURL + endpoint
			if _, exists := uniqueValues[completeURL]; !exists {
				outputFile.WriteString(completeURL + "\n")
				if verboseFlag {
					if colorFlag {
						fmt.Printf("%s%s%s\n", colorGreen, completeURL, colorReset)
					} else {
						fmt.Printf(" %s\n", completeURL)

					}

				}

				uniqueValues[completeURL] = struct{}{}
			}
		}
	}
}
