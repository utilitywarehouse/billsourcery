package bill

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type plainSourceExtractor struct {
	sourceRoot string
	targetDir  string

	inputs  int
	outputs int
}

func (pse *plainSourceExtractor) process(path string) error {
	pse.inputs += 1

	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()

	br := bufio.NewReader(input)

	count := 0

	for {
		prefix, err := br.Peek(4)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch string(prefix) {
		case "TXT,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "TXT,132,")
			spl := strings.SplitN(s, ",", 2)

			c, err := strconv.Atoi(strings.TrimSpace(spl[0]))
			if err != nil {
				return err
			}
			txt := make([]byte, c)
			_, err = io.ReadAtLeast(br, txt, c)
			if err != nil {
				return err
			}

			// We should have a XTX next, which we can discard
			s, _ = br.ReadString('\n')
			if !strings.HasPrefix(s, "XTX,") {
				return fmt.Errorf("expected XTX prefix, but got %s in file %s", strings.TrimSpace(s), path)
			}

			target := pse.targetName(path, count)

			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("error creating directory '%s' : %w", dir, err)
			}

			output, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("error opening output file %s : %w", target, err)
			}
			bw := bufio.NewWriter((output))
			_, err = bw.Write(txt)
			if err := bw.Flush(); err != nil {
				return fmt.Errorf("error flushing output %v : %w", bw, err)
			}
			output.Close()
			if err != nil {
				return fmt.Errorf("error writing to file %v : %w", bw, err)
			}

			count += 1
			pse.outputs += 1

		default:
			_, _ = br.ReadString('\n')
		}

	}

	return nil

}

func (pse *plainSourceExtractor) targetName(path string, count int) string {
	rel, err := filepath.Rel(pse.sourceRoot, path)
	if err != nil {
		panic(fmt.Errorf("path error : %w", err))
	}

	target := filepath.Join(pse.targetDir, rel)

	ext := filepath.Ext(target)

	target = fmt.Sprintf("%s_%d%s", target[0:len(target)-len(ext)], count, ext)

	return target
}
