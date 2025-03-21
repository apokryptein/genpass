package genpass

import (
	"bufio"
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

func PrintLogo() {
	logo := `
  ▄▖
  ▌ █▌▛▌▛▌▀▌▛▘▛▘
  ▙▌▙▖▌▌▙▌█▌▄▌▄▌
        ▌
	`

	notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
	notice(logo)
	// ProcessNewPassFile("test.txt")
}

func GeneratePassphrasewords(wordData map[int][]string, numWords int) string {
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

func ReadWords(filepath string) (map[int][]string, error) {
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

// TODO: add parameter to take desired filename for output
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
		// fmt.Println("[i] ~/.config/genpass Directory Exists")
		if DoesFileExist(homeDir + "/.config/genpass/genpass.lst") {
			// fmt.Println("[i] genpass.lst Exists")
			return true
		} else {
			// fmt.Println("[i] ~.config/genpass/genpass.lst Does Not Exist")
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
		// TODO: clean up -> where to store default list, retrieval, etc
		copyFile("../../configs/genpass.lst", (homeDir + "/.config/genpass/genpass.lst"))
	} else {
		copyFile(wordfile, (homeDir + "/.config/genpass/genpass.lst"))
	}
}

func CopyToClipboard(data string) {
	err := clipboard.WriteAll(data)
	if err != nil {
		fmt.Println("Error copying to clipboard:", err)
		os.Exit(1)
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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
