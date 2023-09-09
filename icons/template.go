package icons

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
)

//go:embed icons.tmpl
var tmpl string

type iconTemplateData struct {
	Provider string
	Icons    []*Icon
}

// writeTemplate generates a template
func writeTemplate(filename string, td iconTemplateData) error {
	f, err := os.OpenFile(filepath.Clean(filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening file (%s): %w", filename, err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	t, err := template.New("icon").Funcs(sprig.FuncMap()).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, td)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	b, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error formatting generated source file")
	}
	_, _ = io.Copy(f, bytes.NewBuffer(b))

	return nil
}
