package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// Qa represents a question-answer pair.
type Qa struct {
	QuizeQuestion string
	QuizAnswer    int
}

func main() {
	fmt.Println("Quiz starts now. Get ready!")

	// Create and save quiz questions to a CSV file.
	quiz := []Qa{
		{
			QuizeQuestion: "What is the addition of 20 + 30?",
			QuizAnswer:    50,
		},
		{
			QuizeQuestion: "What is the addition of 90 - 40?",
			QuizAnswer:    50,
		},
	}
	saveQuizToCSV(quiz, "./Juniperquiz.csv")

	startQuiz("./Juniperquiz.csv")
}

func saveQuizToCSV(quiz []Qa, filename string) {
	// Create or open the CSV file for writing.
	quizFile, err := os.Create(filename)
	checkError(err)
	defer quizFile.Close()

	// Create a CSV writer.
	csvWriter := csv.NewWriter(quizFile)
	defer csvWriter.Flush()

	// Write quiz questions and answers to the CSV file.
	for _, q := range quiz {
		err := csvWriter.Write([]string{q.QuizeQuestion, fmt.Sprintf("%d", q.QuizAnswer)})
		checkError(err)
	}
}
func startQuiz(filename string) {
	records := openAndReadQuizFile(filename)

	score := conductQuiz(records)

	printFinalScore(score)
}

func openAndReadQuizFile(filename string) [][]string {
	quizFile, err := os.Open(filename)
	checkError(err)
	defer quizFile.Close()

	csvReader := csv.NewReader(quizFile)
	records, err := csvReader.ReadAll()
	checkError(err)

	return records
}

func conductQuiz(records [][]string) int {
	var score int
	var questnumber int

	for _, record := range records {
		quizeQuestion := record[0]
		questnumber++

		fmt.Printf("%v. Question: %v \nProvide your answer: ", questnumber, quizeQuestion)

		answerChannel := make(chan string, 1)

		go getUserAnswer(answerChannel)

		select {
		case receivedAnswer := <-answerChannel:
			if receivedAnswer == record[1] {
				score++
			}
		case <-time.After(10 * time.Second):
			fmt.Println("Time's up! Moving to the next question.")
			continue // move to the next question
		}

		if questnumber >= len(records) {
			return score
		}
	}

	return score
}

func getUserAnswer(answerChannel chan string) {
	defer close(answerChannel) // Close the channel when the goroutine exits
	var answer string
	fmt.Scanln(&answer)
	answerChannel <- answer
}

func printFinalScore(score int) {
	switch score {
	case 2:
		fmt.Println("Congratulations, your score is", score)
	case 0:
		fmt.Println("Better luck next time, your score is", score)
	default:
		fmt.Println("You can improve yourself, your score is", score)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
