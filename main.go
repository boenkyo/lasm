package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	cfg      config
	hadError bool
)

type config struct {
	Opcodes map[string]string `json:"opcodes"`
}

func main() {
	var (
		useFile  bool
		filename string
		reader   io.Reader
	)

	cfg = loadConfig()

	switch len(os.Args) {
	case 1:
		useFile = false
		reader = os.Stdin
	case 2:
		useFile = true
	default:
		fmt.Println("Usage: lasm <file>")
		return
	}

	if useFile {
		filename = os.Args[1]
		if !strings.HasSuffix(filename, ".asm") {
			fmt.Println("File must have .asm extension")
			return
		}
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			return
		}
		defer file.Close()
		reader = file
	}

	instructions, tags := parse(reader)
	program := assembleProgram(instructions, tags)

	if hadError {
		return
	}

	hex := convertToHexAndFormat(program)
	if useFile {
		hexFilename := strings.TrimSuffix(filename, ".asm") + ".hex"
		if err := os.WriteFile(hexFilename, []byte(hex), 0644); err != nil {
			fmt.Printf("Error writing to file: %s\n", err)
			return
		}
		fmt.Printf("%d instructions assembled and written to %s.\n\n", len(program), hexFilename)
	} else {
		fmt.Printf("%d instructions assembled:\n\n", len(program))
		fmt.Println("-----")
		fmt.Println(hex)
		fmt.Println("-----")
	}
}

func loadConfig() config {
	var config config
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		panic(err)
	}

	return config
}

func convertToHexAndFormat(program []string) string {
	var hex strings.Builder
	for _, instr := range program {
		binary, err := strconv.ParseInt(instr, 2, 64)
		if err != nil {
			panic(err)
		}
		hex.WriteString(fmt.Sprintf("%04X;\n", binary))
	}

	// Pad with 0s
	for i := len(program); i < 64; i++ {
		hex.WriteString("0000;\n")
	}

	return hex.String()
}

func parse(r io.Reader) ([]string, map[string]int) {
	scanner := bufio.NewScanner(r)

	tags := make(map[string]int)
	var instructions []string
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || isComment(line) {
			continue
		}

		if isTag(line) {
			tagName := line[1:]
			tags[tagName] = lineNum
		} else {
			instructions = append(instructions, line)
			lineNum++
		}
	}

	if err := scanner.Err(); err != nil {
		if !errors.Is(err, io.EOF) {
			panic(err)
		}
	}

	return instructions, tags
}

func assembleProgram(instructions []string, tags map[string]int) []string {
	fmt.Printf("\nAssembling binary:\n\n")
	fmt.Printf("%s\n", strings.Repeat("-", 39))

	var assembled []string
	for line, instr := range instructions {
		program, err := assembleInstruction(instr, tags, line)
		if err != nil {
			fmt.Printf("Error assembling instruction: %s \n %s \n", err, instr)
			hadError = true
			continue
		}
		assembled = append(assembled, program)
	}

	fmt.Printf("%s\n\n", strings.Repeat("-", 39))

	return assembled
}

func assembleInstruction(instruction string, tags map[string]int, line int) (string, error) {
	parts := strings.Fields(instruction)

	if len(parts) < 1 {
		return "", fmt.Errorf("invalid instruction format: %s", instruction)
	}

	opcode, ok := cfg.Opcodes[parts[0]]
	if !ok {
		return "", fmt.Errorf("unknown opcode: %s", parts[0])
	}

	dest, data, err := getDestAndData(parts)
	if err != nil {
		return "", err
	}

	if dest == "" {
		dest = "0"
	}

	if data == "" {
		data = strings.Repeat("0", 8)
	} else {
		data, err = processData(data, tags)
		if err != nil {
			return "", err
		}
	}

	prettyInstruction := fmt.Sprintf("%s %s %s", opcode, dest, data)
	paddedInstruction := fmt.Sprintf("%-20s", instruction)
	fmt.Printf("%d: %s %-13s\n", line, paddedInstruction, prettyInstruction)

	return opcode + dest + data, nil
}

func getDestAndData(parts []string) (dest string, data string, err error) {
	switch len(parts) {
	case 1: // Only opcode
	case 2: // Opcode and either destination or data
		if isDestination(parts[1]) {
			dest, err = processDestination(parts[1])
		} else {
			data = parts[1]
		}
	case 3: // Opcode, destination and data
		dest, err = processDestination(parts[1])
		data = parts[2]
	default:
		err = fmt.Errorf("invalid instruction format: %s", strings.Join(parts, " "))
	}
	return
}

func isDestination(part string) bool {
	return part == "R0" || part == "R1"
}

func processDestination(dest string) (string, error) {
	switch dest {
	case "R0":
		return "0", nil
	case "R1":
		return "1", nil
	default:
		return "", fmt.Errorf("invalid destination: %s", dest)
	}
}

func processData(data string, tags map[string]int) (string, error) {
	if strings.HasPrefix(data, "#") {
		return processTag(data, tags)
	}
	return processBinOrDecData(data)
}

func processTag(data string, tags map[string]int) (string, error) {
	name := data[1:]
	address, ok := tags[name]
	if !ok {
		return "", fmt.Errorf("unknown tag: %s", name)
	}
	return fmt.Sprintf("%08b", address), nil
}

func processBinOrDecData(data string) (string, error) {
	if strings.HasPrefix(data, "0b") {
		// Data is in binary format
		data = data[2:]
		if len(data) != 8 {
			return "", fmt.Errorf("binary data should be 8 bits long: %s", data)
		}
		return data, nil
	}

	// Data is in decimal format
	decimal, err := strconv.Atoi(data)
	if err != nil {
		return "", fmt.Errorf("invalid decimal data: %s", data)
	}
	return fmt.Sprintf("%08b", decimal), nil
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "//")
}

func isTag(line string) bool {
	return strings.HasPrefix(line, "#")
}
