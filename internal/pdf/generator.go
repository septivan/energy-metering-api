package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/signintech/gopdf"

	"energy-metering-api/internal/billing"
	"energy-metering-api/internal/config"
)

type Generator struct {
	outDir string
}

func NewGenerator(cfg *config.Config) *Generator {
	dir := cfg.PDFOutputDir
	_ = os.MkdirAll(dir, 0o755)
	return &Generator{outDir: dir}
}

func (g *Generator) GenerateInvoice(ctx context.Context, preview *billing.BillingPreviewResponse) (string, error) {
	fname := fmt.Sprintf("invoice-%s-%s.pdf", preview.ClientID, time.Now().Format("20060102150405"))
	path := filepath.Join(g.outDir, fname)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// Set font - needs to be loaded first
	// For basic usage without font file, we can skip SetFont or use AddTTFFont
	// Using simple text without font for now

	pdf.SetX(50)
	pdf.SetY(50)
	err := pdf.Cell(nil, fmt.Sprintf("Client: %s", preview.ClientID))
	if err != nil {
		return "", err
	}

	pdf.Br(20)
	err = pdf.Cell(nil, fmt.Sprintf("Period: %s - %s", preview.Period.Start, preview.Period.End))
	if err != nil {
		return "", err
	}

	pdf.Br(20)
	err = pdf.Cell(nil, fmt.Sprintf("Total Usage: %.3f kWh", preview.UsageKwh))
	if err != nil {
		return "", err
	}

	pdf.Br(20)
	err = pdf.Cell(nil, fmt.Sprintf("Min Reading: %.3f kWh", preview.Calculation.MinKwh))
	if err != nil {
		return "", err
	}

	pdf.Br(20)
	err = pdf.Cell(nil, fmt.Sprintf("Max Reading: %.3f kWh", preview.Calculation.MaxKwh))
	if err != nil {
		return "", err
	}

	// Write to file
	if err := pdf.WritePdf(path); err != nil {
		return "", err
	}

	return path, nil
}
