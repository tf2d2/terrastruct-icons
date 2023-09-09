package icons

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// Icon contains all relevant details to use an icon
type Icon struct {
	Cloud string
	Title string
	URL   string
}

const (
	sourceURL  = "https://icons.terrastruct.com"
	outputDir  = "output"
	outputFile = "icons.csv"
)

var (
	// Valid cloud providers
	clouds    = []string{"aws", "azure", "gcp"}
	cloudsRgx = regexp.MustCompile(`(aws|azure|gcp)`)
	// Regular expression to match Unicode escape sequences (\uXXXX)
	escapeRgx = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
)

// Generate generates a csv output file with icon details
func Generate() error {
	log.Println("Generate icon details")

	checkOutputDirectories()

	c := colly.NewCollector()
	c.OnError(func(r *colly.Response, err error) {
		log.Fatalf("error escraping %s: %s", r.Request.URL, err.Error())
	})

	icons := make(map[string][]*Icon)
	c.OnHTML("div", func(e *colly.HTMLElement) {
		if e.Attr("class") == "icon" {
			unescaped := getUnescaped(e.Attr("onclick"))
			link := strings.TrimSuffix(strings.TrimPrefix(unescaped, "clickIcon(\""), "\")")
			match := cloudsRgx.MatchString(link)
			if match {
				i := Icon{
					Cloud: strings.ToUpper(strings.Split(link, "%")[0]),
					Title: e.Attr("data-search"),
					URL:   fmt.Sprintf("%s/%s", sourceURL, link),
				}
				icons[i.Cloud] = append(icons[i.Cloud], &i)
			}
		}
	})
	_ = c.Visit(sourceURL)

	// generate csv
	err := writeCSV(icons)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// generate json
	for k, v := range icons {
		cloud := strings.ToLower(k)
		path := filepath.Join(outputDir, cloud, fmt.Sprintf("%s.json", cloud))
		f, err := os.OpenFile(filepath.Clean(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("error opening file %s.json: %s", k, err.Error())
		}
		e := json.NewEncoder(f)
		e.SetEscapeHTML(false)
		e.SetIndent("", " ")
		err = e.Encode(v)
		if err != nil {
			log.Fatalf("error encoding %s icon data: %s", path, err.Error())
		}
		err = f.Close()
		if err != nil {
			log.Fatalf("error closing %s file handler: %s", path, err.Error())
		}
		log.Printf("Output file: %s", path)
	}

	// generate templates
	for k, v := range icons {
		td := iconTemplateData{
			Provider: k,
			Icons:    v,
		}
		cloud := strings.ToLower(k)
		path := filepath.Join(outputDir, cloud, fmt.Sprintf("%s.go", cloud))
		err := writeTemplate(path, td)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	log.Printf("Output file: %s", outputFile)

	return nil
}

func getUnescaped(escaped string) string {
	// Replace the matched escape sequences with their unescaped versions
	unescapedString := escapeRgx.ReplaceAllStringFunc(escaped, func(match string) string {
		// Extract the hex code from the match
		hexCode := match[2:] // Removing the "\u" prefix
		unicodeValue, err := strconv.ParseInt(hexCode, 16, 32)
		if err != nil {
			log.Fatalf("error unescaping string %s: %s", escaped, err.Error())
		}
		// Convert the Unicode value to a string and return
		return string(rune(unicodeValue))
	})

	return unescapedString
}

func writeCSV(icons map[string][]*Icon) error {
	f, err := os.Create(filepath.Join(outputDir, outputFile))
	if err != nil {
		return fmt.Errorf("failed to create %s: %s", outputFile, err.Error())
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	var header []string
	iconType := Icon{
		Cloud: "",
		Title: "",
		URL:   sourceURL,
	}
	t := reflect.TypeOf(iconType)
	for i := 0; i < t.NumField(); i++ {
		header = append(header, t.Field(i).Name)
	}
	err = w.Write(header)
	if err != nil {
		return fmt.Errorf("error writing header: %s", err.Error())
	}

	// rows
	for _, v := range icons {
		for _, iv := range v {
			icon := []string{
				iv.Cloud,
				iv.Title,
				iv.URL,
			}
			err := w.Write(icon)
			if err != nil {
				return fmt.Errorf("failed to write csv data: %s", err.Error())
			}
		}
	}

	return nil
}

func checkOutputDirectories() {
	for _, c := range clouds {
		path := filepath.Join(outputDir, c)
		_, err := os.Stat(path)
		switch {
		case os.IsNotExist(err):
			err = os.MkdirAll(path, 0750)
			if err != nil {
				log.Fatalf("error creating output directory: %s", err.Error())
			}
			log.Println("created output directory", path)
		case err != nil:
			log.Fatalf("error checking output directory: %s", err.Error())
		}
	}
}
