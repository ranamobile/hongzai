package lib

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type Score struct {
	Name  string
	Count int
}

type ScoreList struct {
	Scores []Score
}

func (scorelist ScoreList) Len() int {
	return len(scorelist.Scores)
}

func (scorelist ScoreList) Less(i, j int) bool {
	return scorelist.Scores[i].Count < scorelist.Scores[j].Count
}

func (scorelist ScoreList) Swap(i, j int) {
	scorelist.Scores[i], scorelist.Scores[j] = scorelist.Scores[j], scorelist.Scores[i]
}

func (scorelist *ScoreList) Read(filename string) {
	handler, err := os.Open(filename)
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	defer handler.Close()

	reader := csv.NewReader(handler)
	reader.FieldsPerRecord = 2

	for {
		line, err := reader.Read()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		count, err := strconv.Atoi(line[1])
		if err != nil {
			log.Println("Read error:", err)
			continue
		}
		scorelist.Update(line[0], count)
	}
}

func (scorelist *ScoreList) Write(filename string) {
	handler, err := os.Create(filename)
	if err != nil {
		log.Println("Write error:", err)
		return
	}
	defer handler.Close()

	writer := csv.NewWriter(handler)
	defer writer.Flush()

	log.Println("Scores:", len(scorelist.Scores))
	for _, score := range scorelist.Scores {
		log.Println("Writing score:", score.Name, score.Count)
		writer.Write([]string{score.Name, strconv.Itoa(score.Count)})
	}
	writer.Flush()
}

func (scorelist *ScoreList) Find(name string) int {
	for index, score := range scorelist.Scores {
		if score.Name == name {
			return index
		}
	}
	return -1
}

func (scorelist *ScoreList) Update(name string, count int) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.Scores)
		scorelist.Scores = append(scorelist.Scores, Score{name, count})
	} else {
		scorelist.Scores[index].Count = count
	}
	return scorelist.Scores[index]
}

func (scorelist *ScoreList) Increment(name string) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.Scores)
		scorelist.Scores = append(scorelist.Scores, Score{name, 1})
	} else {
		scorelist.Scores[index].Count++
	}
	return scorelist.Scores[index]
}

func (scorelist *ScoreList) Decrement(name string) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.Scores)
		scorelist.Scores = append(scorelist.Scores, Score{name, -1})
	} else {
		scorelist.Scores[index].Count--
	}
	return scorelist.Scores[index]
}
