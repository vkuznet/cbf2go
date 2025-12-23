// cbf.go
// Minimal but correct CBF (CIF Binary File) reader for x-CBF_BYTE_OFFSET images
// Ported to match fabio.open(...).data behavior

package cbf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ReadCBF(path string, verbose int) ([]int32, int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}

	// ------------------------------------------------------------
	// Locate binary starter
	// ------------------------------------------------------------
	starter := []byte{0x0c, 0x1a, 0x04, 0xd5}
	pos := bytes.Index(data, starter)
	if pos < 0 {
		return nil, 0, 0, fmt.Errorf("CBF binary starter not found")
	}

	headerText := string(data[:pos])
	binaryStart := pos + len(starter)

	// ------------------------------------------------------------
	// Parse header text
	// ------------------------------------------------------------
	header := parseCBFHeader(headerText)
	if verbose > 0 {
		fmt.Println("CBF header")
		for k, v := range header {
			fmt.Printf("%v: %v\n", k, v)
		}
	}

	// Dimensions (these ARE present in your file)
	w, err := strconv.Atoi(header["X-Binary-Size-Fastest-Dimension"])
	if err != nil {
		return nil, 0, 0, err
	}
	h, err := strconv.Atoi(header["X-Binary-Size-Second-Dimension"])
	if err != nil {
		return nil, 0, 0, err
	}

	nElem, err := strconv.Atoi(header["X-Binary-Number-of-Elements"])
	if err != nil {
		return nil, 0, 0, err
	}

	if nElem != w*h {
		return nil, 0, 0, fmt.Errorf("element mismatch: %d vs %d", nElem, w*h)
	}

	// ------------------------------------------------------------
	// Extract EXACT binary payload
	// ------------------------------------------------------------
	binSize, err := strconv.Atoi(header["X-Binary-Size"])
	if err != nil {
		return nil, 0, 0, err
	}

	if binaryStart+binSize > len(data) {
		return nil, 0, 0, fmt.Errorf("binary data truncated")
	}

	binaryData := data[binaryStart : binaryStart+binSize]

	// ------------------------------------------------------------
	// Decode BYTE_OFFSET
	// ------------------------------------------------------------
	pixels, err := decByteOffsetFabio(binaryData, nElem)
	if err != nil {
		return nil, 0, 0, err
	}
	if verbose > 0 {
		fmt.Println("### first 10 pixels", pixels[:10])
	}

	return pixels, w, h, nil
}

// ------------------------------------------------------------
// BYTE_OFFSET decoder (Fabio-compatible)
// ------------------------------------------------------------
func decByteOffsetFabio(raw []byte, size int) ([]int32, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty byte_offset stream")
	}

	out := make([]int32, size)
	r := bytes.NewReader(raw)

	// ------------------------------------------------------------
	// FIRST pixel: ABSOLUTE int8  (Pilatus / FabIO behavior)
	// ------------------------------------------------------------
	var first int8
	if err := binary.Read(r, binary.LittleEndian, &first); err != nil {
		return nil, err
	}
	out[0] = int32(first)

	// ------------------------------------------------------------
	// Remaining pixels: BYTE_OFFSET deltas
	// ------------------------------------------------------------
	for i := 1; i < size; i++ {
		var d8 int8
		if err := binary.Read(r, binary.LittleEndian, &d8); err != nil {
			return nil, fmt.Errorf("byte_offset truncated at pixel %d", i)
		}

		delta := int32(d8)

		if d8 == -128 {
			var d16 int16
			if err := binary.Read(r, binary.LittleEndian, &d16); err != nil {
				return nil, err
			}
			delta = int32(d16)

			if d16 == -32768 {
				var d32 int32
				if err := binary.Read(r, binary.LittleEndian, &d32); err != nil {
					return nil, err
				}
				delta = d32
			}
		}

		out[i] = out[i-1] + delta
	}

	return out, nil
}

// ------------------------------------------------------------
// CBF parsing helpers
// ------------------------------------------------------------

var (
	binaryMarker = []byte("--CIF-BINARY-FORMAT-SECTION--")
	starter      = []byte{0x0c, 0x1a, 0x04, 0xd5}
)

func readCBFSections(r io.Reader) (map[string]string, []byte, error) {
	br := bufio.NewReader(r)
	var headerBuf bytes.Buffer

	// Read ASCII header until binary marker
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			return nil, nil, err
		}
		if bytes.Contains(line, binaryMarker) {
			headerBuf.Write(line)
			break
		}
		headerBuf.Write(line)
	}

	header := parseCBFHeader(headerBuf.String())

	// Read binary header until starter
	var binHeader bytes.Buffer
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil, nil, err
		}
		binHeader.WriteByte(b)
		if binHeader.Len() >= len(starter) &&
			bytes.Equal(binHeader.Bytes()[binHeader.Len()-len(starter):], starter) {
			break
		}
	}

	// Now read binary payload
	size, err := strconv.Atoi(header["X-Binary-Size"])
	if err != nil {
		return nil, nil, err
	}

	binaryData := make([]byte, size)
	_, err = io.ReadFull(br, binaryData)
	if err != nil {
		return nil, nil, err
	}

	return header, binaryData, nil
}

func parseCBFHeader(txt string) map[string]string {
	h := make(map[string]string)

	lines := strings.Split(txt, "\n")
	for i := 0; i < len(lines); i++ {
		l := strings.TrimSpace(lines[i])

		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}

		// Skip CIF control lines
		if l == "loop_" || l == ";" {
			continue
		}

		// key: value
		if strings.Contains(l, ":") {
			p := strings.SplitN(l, ":", 2)
			k := strings.TrimSpace(p[0])
			v := strings.Trim(strings.TrimSpace(p[1]), "\"")
			h[k] = v
			continue
		}

		// key value
		fields := strings.Fields(l)
		if len(fields) >= 2 && strings.HasPrefix(fields[0], "X-Binary-") {
			h[fields[0]] = strings.Trim(fields[1], "\"")
			continue
		}

		// CIF-style tags (_array_data.data etc)
		if strings.HasPrefix(l, "_") {
			// value might be on same line or next line
			if len(fields) > 1 {
				h[fields[0]] = fields[1]
			} else if i+1 < len(lines) {
				next := strings.TrimSpace(lines[i+1])
				if next != "" && !strings.HasPrefix(next, "_") {
					h[fields[0]] = next
					i++
				}
			}
		}
	}

	return h
}
