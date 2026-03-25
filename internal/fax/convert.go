package fax

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertToTIFF(inputPath, outputDir string) (string, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	outName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)) + ".tiff"
	outPath := filepath.Join(outputDir, outName)

	// Check input file size
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return "", fmt.Errorf("stat input: %w", err)
	}
	largeFile := inputInfo.Size() > 100*1024

	switch ext {
	case ".pdf":
		if largeFile {
			if err := convertPDFLowRes(inputPath, outPath); err != nil {
				return "", err
			}
		} else {
			if err := convertPDF(inputPath, outPath); err != nil {
				return "", err
			}
		}
	case ".png", ".jpg", ".jpeg":
		if largeFile {
			if err := convertImageLowRes(inputPath, outPath); err != nil {
				return "", err
			}
		} else {
			if err := convertImage(inputPath, outPath); err != nil {
				return "", err
			}
		}
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	return outPath, nil
}

func convertPDF(input, output string) error {
	cmd := exec.Command("gs",
		"-q", "-dNOPAUSE", "-dBATCH",
		"-sDEVICE=tiffg3",
		"-r204x196",
		"-dPDFFitPage",
		"-g1728x2292",
		"-sOutputFile="+output,
		input,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ghostscript: %w (output: %s)", err, string(out))
	}
	return nil
}

func convertPDFLowRes(input, output string) error {
	cmd := exec.Command("gs",
		"-q", "-dNOPAUSE", "-dBATCH",
		"-sDEVICE=tiffg3",
		"-r204x98",
		"-dPDFFitPage",
		"-g1728x1146",
		"-sOutputFile="+output,
		input,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ghostscript low-res: %w (output: %s)", err, string(out))
	}
	return nil
}

func convertImage(input, output string) error {
	cmd := exec.Command("convert",
		input,
		"-background", "white",
		"-alpha", "remove",
		"-alpha", "off",
		"-gravity", "center",
		"-extent", "1728x2292",
		"-grayscale", "Rec709Luminance",
		"-contrast-stretch", "2%x2%",
		"-sharpen", "0x0.5",
		"-dither", "FloydSteinberg",
		"-colors", "2",
		"-density", "204x196",
		"-units", "PixelsPerInch",
		"-compress", "Fax",
		"-type", "bilevel",
		output,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick: %w (output: %s)", err, string(out))
	}
	return nil
}

func convertImageLowRes(input, output string) error {
	cmd := exec.Command("convert",
		input,
		"-background", "white",
		"-alpha", "remove",
		"-alpha", "off",
		"-resize", "1600x1050",
		"-gravity", "center",
		"-extent", "1728x1146",
		"-grayscale", "Rec709Luminance",
		"-contrast-stretch", "2%x2%",
		"-sharpen", "0x0.5",
		"-dither", "FloydSteinberg",
		"-colors", "2",
		"-density", "204x98",
		"-units", "PixelsPerInch",
		"-compress", "Fax",
		"-type", "bilevel",
		output,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick low-res: %w (output: %s)", err, string(out))
	}
	return nil
}
