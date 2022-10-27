package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// a quiz is comprised of question objects each having the question itself, the correct answer to the question, and a check
// for the correct response
type question struct {
	question             string
	answer               string
	correctResponseGiven bool
}

func main() {
	// setup flag for -csv to allow a user-specified list of questions and answers to be used
	csvPtr := flag.String("csv", "problems.csv", "csv file in the format of 'question,answer'")
	// setup flag for -limit to allow a user-specified time limit to answer the questions
	limitPtr := flag.Int("limit", 30, "the time limit for the quiz in seconds")
	// Parse() parses all command-line flags in os.Args[1:]
	flag.Parse()

	// open file for read only
	quizFile, err := os.Open(*csvPtr)
	// check for error opening file
	if err != nil {
		// if error, print message, log the error, and exit
		fmt.Printf("failed to open CSV file %s\n", *csvPtr)
		log.Fatal(err)
	}
	// if no error, defer the close of the file
	defer quizFile.Close()

	// declare a new reader which can read the CSV file
	r := csv.NewReader(quizFile)
	// using the opened file, create a slice of question objects
	quiz := createQuiz(r)

	// declare a time limit based on the flag value
	timeLimit := time.Second * time.Duration(*limitPtr)

	// prompt user to begin test
	fmt.Print("Please press enter to begin quiz...")
	// wait until enter is pressed - Scanln() scans until a new line
	fmt.Scanln()

	// declare channel to indicate the quiz is complete
	done := make(chan bool)
	// declare variable to count correct answers
	correctAnswers := 0

	go runQuiz(quiz, &correctAnswers, done)
	for {
		select {
		// in this case the time limit has been reached
		case <-time.After(timeLimit):
			fmt.Printf("\nYou scored %v out of %v.\n", correctAnswers, len(quiz))
			return
		// in this case, the quiz has been completed in the allowed time
		case <-done:
			fmt.Printf("You scored %v out of %v.\n", correctAnswers, len(quiz))
			return
		}
	}
}

// createQuiz() reads each line of the CSV and initializes a slice of question objects
func createQuiz(r *csv.Reader) []question {
	// quizzes are assumed to be no greater than 100 questions - the underlying array will have a capacity of 100 to accomodate
	quiz := make([]question, 0, 100)
	for {
		// read from the CSV file
		record, err := r.Read()
		if err != nil {
			// check if the record had an unexpected number of fields
			if err == csv.ErrFieldCount {
				fmt.Println("record from quiz file had an unexpected number of fields")
				log.Fatal(err)
			}
			// check if the end of the file has been reached
			if err == io.EOF {
				// break out of for loop
				break
			}
			// otherwise log error and exit
			log.Fatal(err)
		}
		// populate the question and the answer from the record read from the CSV, appending the new question object to the
		// slice - strings.TrimSpace will remove any leading or trailing white spaces around the answer
		quiz = append(quiz, question{question: record[0], answer: strings.TrimSpace(record[1]), correctResponseGiven: false})
	}
	return quiz
}

// runQuiz() starts the quiz and updates the count of questions answered correctly
func runQuiz(quiz []question, correctAnswers *int, done chan bool) {
	// declare new scanner to read from the standard input
	s := bufio.NewScanner(os.Stdin)

	for i, quizQuestion := range quiz {
		// print the question
		fmt.Printf("Problem #%v: %s = ", i+1, quizQuestion.question)
		// scan for an answer
		s.Scan()
		// compare answer given to the correct answer - removing any leading or trailing white spaces as we can assume all
		// answers will be a single word or number
		if quizQuestion.answer == strings.TrimSpace(s.Text()) {
			// update correctResponseGiven - keeps track of which questions were answered correctly and which were answered
			// incorrectly should we want to list which questions were answered incorrectly in the future
			quiz[i].correctResponseGiven = true
			*correctAnswers++
		}
	}
	done <- true
	return
}
