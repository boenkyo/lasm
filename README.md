# LASM (Lejon Assembly Language)

LASM is a basic language and assembler for an MCU that students at Lund University construct in the course EITF65.
The code is assembled and converted to a hexadecimal format to be compatible with the software the university provides for programming the MCU.

## Compiling

Make sure you have the latest version of [go](https://go.dev/dl/) installed and run

`make`

## Usage

### Assemble from a file
`lasm <input file>`

The output will be written to a `.hex` file with the same name and in the same directory as the input file.

### Assemble from standard input
`lasm`

Write the instructions line by line and press `Ctrl + D` to assemble them.

### Configuration

The names and opcodes of the instructions can be configured in `config.json`.

## Examples

The following program is a simple loop that loads the value 10 into register R0, decrements R0 until it reaches 0, and then ends the loop.

Note the usage of `// comments` and `#labels` to avoid having to explicitly write the instruction address.
```
// Load 10 into R0
LOD R0 10

// Loop until R0 is 0
#loop
SUB R0 1
BRZ R0 #end
BRN #loop

#end
BRN #end
```

## Alternatives

[ALP](https://github.com/julius-andreasson/ALP/tree/main) is another assembler written in Python by students at Lund University.