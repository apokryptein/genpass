package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	// Print the logo
	printLogo()

	// Define command-line flags
	wordlist := flag.String("wordlist", "", "Path to the wordlist file")
	numWords := flag.Int("num", 4, "Number of words in the passphrase")
	process := flag.Bool("process", false, "Process new wordlist for use with GenPass")
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
	good := CheckConfig(*wordlist, homeDir, isDefault)

	if good {
		fmt.Printf("%s Looks like your config is good to go.\n\n", yellow("[i]"))
	}

	// parse wordlist into map[int][]string data structure
	wordData, _ := readWords(*wordlist)

	// get passphrase
	passPhrase := generatePassphrasewords(wordData, *numWords)

	// customize print function
	passPrint := color.New(color.Bold, color.FgMagenta, color.BgGreen).SprintFunc()
	fmt.Printf("Your new password: %s\n", passPrint(passPhrase))

	// copy to clipboard if desired
	if *copyToClip {
		CopyToClipboard(passPhrase)
		color.Red("Your new password has been copied to the clipboard.")
	}
}

func printLogo() {
	logo := `
	  __             __                
	 /              /  |               
	( __  ___  ___ (___| ___  ___  ___ 
	|   )|___)|   )|    |   )|___ |___ 
	|__/ |__  |  / |    |__/| __/  __/
	`

	notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
	notice(logo)
	//ProcessNewPassFile("test.txt")
}

func generatePassphrasewords(wordData map[int][]string, numWords int) string {
	// additional characters to append (for now)
	specials := [6]string{"@", "&", "$", "!", "#", "?"}
	nums := [10]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	// generate random seed using time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// string array to store chosen words for passphrase
	passphraseWords := make([]string, numWords+3)

	// cases object to capitalize first letter of each chosen word
	caser := cases.Title(language.English)

	// length of wordData map([int][]string) -> number of keys
	dataLen := len(wordData)

	// get key values and store in array
	keys := make([]int, 0, len(wordData))
	for k := range wordData {
		keys = append(keys, k)
	}

	// choose number of desired words for passphrase and store in passphraseWords array
	for i := 0; i < numWords; i++ {
		wordLen := keys[r.Intn(dataLen)]
		wordChoice := r.Intn(len(wordData[wordLen]))
		passphraseWords[i] = caser.String(wordData[wordLen][wordChoice])
	}

	for i := numWords; i < numWords+3; i++ {
		if i == numWords {
			passphraseWords[i] = specials[r.Intn(len(specials))]
			continue
		}
		passphraseWords[i] = nums[r.Intn(len(nums))]
	}

	// join them and return
	return strings.Join(passphraseWords, "")
}

func readWords(filepath string) (map[int][]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[int][]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// split line into array of two values: [num word]
		vals := strings.Fields(scanner.Text())

		if len(vals) != 2 {
			fmt.Println("[!] Wordlist is not properly formatted. Exiting...")
			os.Exit(1)
		}

		// convert number string to int
		strconv.Atoi(vals[0])

		// get wordlength value and conver to int
		wordLen, err := strconv.Atoi(vals[0])
		checkErr(err)

		// check if length is already a key in our map
		_, ok := data[wordLen]

		if ok {
			// add word to existing array in map
			data[wordLen] = append(data[wordLen], vals[1])
		} else {
			// val does not yet exist, make it
			data[wordLen] = []string{vals[1]}
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func ProcessNewPassFile(wordfile string) {
	// Open provided file for parsing
	file, err := os.Open(wordfile)

	checkErr(err)

	// Create new formatted file for use with tool
	out_file, err := os.Create("updated_wordfile.txt")

	checkErr(err)

	defer file.Close()
	defer out_file.Close()

	// NOTE: This will only work for wordfiles with fewer than 65,535 lines
	// TODO: Implement check function to see how many lines in file to adjust accordingly
	scanner := bufio.NewScanner(file)

	// get char count for each word (one word per line)
	// output each line to new file containing genpass format -> [0-9]\s[a-z]{3,} (each word in list should be a minimum of 3 characters)
	for scanner.Scan() {
		char_count := utf8.RuneCountInString(scanner.Text())
		fmt.Fprintln(out_file, (strconv.Itoa(char_count) + " " + strings.ToLower(scanner.Text())))
	}

}

func DoesDirExist(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DoesFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		return false
	}
}

func CheckConfig(wordfile string, homeDir string, isDefault bool) bool {
	val, _ := DoesDirExist(homeDir + "/.config/genpass")

	if val {
		//fmt.Println("[i] ~/.config/genpass Directory Exists")
		if DoesFileExist(homeDir + "/.config/genpass/genpass.lst") {
			//fmt.Println("[i] genpass.lst Exists")
			return true
		} else {
			//fmt.Println("[i] ~.config/genpass/genpass.lst Does Not Exist")
			CreateConfig(wordfile, homeDir, isDefault)
		}

	} else {
		yellow := color.New(color.FgYellow).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Printf("%s Config directory not present. Creating it for you here: %s\n", yellow("[i]"), cyan(homeDir+"/.config/genpass"))
		err := os.Mkdir(homeDir+"/.config/genpass", 0755)
		checkErr(err)
		CreateConfig(wordfile, homeDir, isDefault)
	}
	return false
}

func CreateConfig(wordfile string, homeDir string, isDefault bool) {
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s Wordlist config not present. Creating it for you here: %s\n\n", yellow("[i]"), cyan(homeDir+"/.config/genpass/genpass.lst"))

	if isDefault {
		copyFile("genpass.lst", (homeDir + "/.config/genpass/genpass.lst"))
	} else {
		copyFile(wordfile, (homeDir + "/.config/genpass/genpass.lst"))
	}

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func copyFile(src string, dst string) {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = os.WriteFile(dst, data, 0644)
	checkErr(err)
}

func CopyToClipboard(data string) {
	err := clipboard.WriteAll(data)
	if err != nil {
		fmt.Println("Error copying to clipboard:", err)
		os.Exit(1)
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
