package deploy

import "fmt"

func PrintHeaderToConsole(text string) {
	fmt.Println("")
	fmt.Println("---------------------------------")
	fmt.Printf("%s\n", text)
	fmt.Println("---------------------------------")
}

func Obfuscate(s string) string {
	out := []rune(s)
	for i := 3; i < len(s); i++ {
		out[i] = '*'
	}
	return string(out)
}
