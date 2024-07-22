package main

import (
	"flag"
	"fmt"
	"os"

	gp "github.com/apokryptein/genpass"
	"github.com/fatih/color"
)

func main() {
	// Print the logo
	gp.PrintLogo()

	// Define command-line flags
	wordlist := flag.String("w", "", "Path to the wordlist file")
	numWords := flag.Int("n", 4, "Number of words in the passphrase")
	process := flag.Bool("p", false, "Process new wordlist for use with GenPass")
	copyToClip := flag.Bool("c", false, "Copy new password to clipboard")
	flag.Parse()

	// get user's home directory
	homeDir, _ := os.UserHomeDir()
	isDefault := false

	// set notice print color
	yellow := color.New(color.FgYellow).SprintFunc()

	// Check if wordlist path is provided
	if !(isFlagPassed(*wordlist)) {
		fmt.Printf("%s Using default wordlist\n", yellow("[i]"))
		*wordlist = homeDir + "/.config/genpass/genpass.lst"
		isDefault = true
	} else if *process && !(isFlagPassed(*wordlist)) {
		fmt.Println("Please provide a path to the wordlist you would like to process using the -wordlist flag.")
		return
	}

	// check if ~/.config/genpass & ~/config/genpass/genpass.lst are present
	good := gp.CheckConfig(*wordlist, homeDir, isDefault)

	if good {
		fmt.Printf("%s Looks like your config is good to go.\n\n", yellow("[i]"))
	}

	// parse wordlist into map[int][]string data structure
	wordData, _ := gp.ReadWords(*wordlist)

	// get passphrase
	passPhrase := gp.GeneratePassphrasewords(wordData, *numWords)

	// customize print function
	passPrint := color.New(color.Bold, color.FgMagenta).SprintFunc()
	fmt.Printf("Your new password: %s\n", passPrint(passPhrase))

	// copy to clipboard if desired
	if *copyToClip {
		gp.CopyToClipboard(passPhrase)
		color.Red("Your new password has been copied to the clipboard.")
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
