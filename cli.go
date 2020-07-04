package main

import (
	"fmt"
	_ "github.com/confluentinc/confluent-kafka-go/kafka"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"text/template"
	"time"
)

type Monad struct {
	MonadInput  MonadInput
	CalcResults []CalcResult
}
type MonadInput struct {
	MiljonerTransaktioner int
	MiljonerSumma         int
}
type Betalning struct {
	Id      uint64
	Datum   string
	Valuta  int
	LandKod int
	Belopp  float32
}

type CalcResult struct {
	Betalningar []Betalning
	// OutputFiles  []*os.File
	OutputFile *os.File
}

func main() {
	start := time.Now()

	const (
		KafkaRender = "KAFKARENDER"
		FileRender  = "FILERENDER"
	)
	renderMode := FileRender
	temp, err := ioutil.ReadFile("templates/iso200200/example_camt.053_swedish_account_statement.xml")
	if err != nil {
		panic(err)
	}
	var t, _ = template.New("bla").Parse(string(temp))


	yearDistribution := [12]Monad{{MonadInput: MonadInput{
		MiljonerTransaktioner: 10,
		MiljonerSumma:         10,
	}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 15,
			MiljonerSumma:         150,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 15,
			MiljonerSumma:         150,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 20,
			MiljonerSumma:         200,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 20,
			MiljonerSumma:         200,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 25,
			MiljonerSumma:         25,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 40,
			MiljonerSumma:         40,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 40,
			MiljonerSumma:         40,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 40,
			MiljonerSumma:         40,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 20,
			MiljonerSumma:         20,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 20,
			MiljonerSumma:         20,
		}, CalcResults: nil},
		{MonadInput: MonadInput{
			MiljonerTransaktioner: 50,
			MiljonerSumma:         50,
		}, CalcResults: nil}}

	const cpuCores int = 20
	cpuCoresPerMonth := calcCoresPerMonad(cpuCores,len(yearDistribution))

	totTrans := int(0)
	totSum := int(0)

	// Calculate

	// for _, monad := range yearDistribution
	var wgCalc sync.WaitGroup

	for m := 0; m < len(yearDistribution); m++ {
		trans := yearDistribution[m].MonadInput.MiljonerTransaktioner * 1000000
		sum := yearDistribution[m].MonadInput.MiljonerSumma

		totTrans = totTrans + trans
		totSum = totSum + sum
		BetalningarDennaMonad := make([]Betalning, trans)

		yearDistribution[m].CalcResults = make([]CalcResult, cpuCoresPerMonth)

		istart := 0
		istop := 0
		for j := 0; j < cpuCoresPerMonth; j++ {
			istart, istop = getSliceIndexes(j, trans, cpuCoresPerMonth)
			CalcResult2 := CalcResult{
				Betalningar: BetalningarDennaMonad[istart:istop+1],
			}

			yearDistribution[m].CalcResults[j] = CalcResult2
			stor := len(CalcResult2.Betalningar)
			wgCalc.Add(1)
			go skapaBetalningar(j, stor, sum/cpuCoresPerMonth, &yearDistribution[m].CalcResults[j].Betalningar, &wgCalc)
		}
	}
	wgCalc.Wait()

	elapsed := time.Since(start)
	fmt.Printf("In memory generated total of %d MILLIONEN transactions  . Generation Speed= %d of Transactions per Second \n", totSum,int32 (float64(totTrans) / elapsed.Seconds()))



	// Render out
	var wgRender sync.WaitGroup

	if renderMode == FileRender {
		for m := 0; m < len(yearDistribution); m++ {
			for j := 0; j < cpuCoresPerMonth; j++ {
				f, err := os.Create("/tmp/iso200200" + strconv.Itoa(m) + "_" + strconv.Itoa(j))
				if err != nil {
					log.Println("create file: ", err)
					return
				}
				yearDistribution[m].CalcResults[j].OutputFile = f
				wgRender.Add(1)
				go JulleBajsarBetalningar(yearDistribution[m].CalcResults[j].OutputFile, t, &yearDistribution[m].CalcResults[j], &wgRender)
			}
		}
	}

	if renderMode == KafkaRender {
		for m := 0; m < len(yearDistribution); m++ {
			for j := 0; j < cpuCoresPerMonth; j++ {
				f, err := os.Create("/tmp/output" + strconv.Itoa(m) + "_" + strconv.Itoa(j))
				if err != nil {
					log.Println("create file: ", err)
					return
				}
				yearDistribution[m].CalcResults[j].OutputFile = f
				wgRender.Add(1)
				go JulleBajsarBetalningar(yearDistribution[m].CalcResults[j].OutputFile, t, &yearDistribution[m].CalcResults[j], &wgRender)
			}
		}
	}

	wgRender.Wait()
	fmt.Printf("totTrans %d", totTrans)
	elapsed2 := time.Since(start)

	fmt.Printf("Transactions per Second %d \n", int32(float64(totTrans) / elapsed2.Seconds()))



	log.Printf(" took %s to generate ", elapsed2)
}

func skapaBetalningar(core, antal, summa int, betalningarPek *[]Betalning, wg *sync.WaitGroup) {
	defer wg.Done()
	var betalningar []Betalning = *betalningarPek

	sekvens := strconv.Itoa(core)
	for i := 0; i < antal; i++ {
		betalningar[i].Id = rand.Uint64()
		betalningar[i].Belopp = rand.Float32() * 2000
		betalningar[i].Datum = "2020-" + sekvens + "-num-" + strconv.Itoa(i)
		betalningar[i].Valuta = rand.Intn(240)
		betalningar[i].LandKod = rand.Intn(240)
	}

}

func JulleBajsarBetalningar(filen *os.File, t *template.Template, calcResult *CalcResult, wg *sync.WaitGroup) {
	defer wg.Done()

	e := t.Execute(filen, calcResult)
	if e != nil {
		log.Println("create file: ", e)
		return
	}

}

func getSliceIndexes(segnr, trans, cores int) (start, stop int) {
	segsize := trans / cores
	start = segnr * segsize
	stop = start + segsize - 1
	if trans-stop < 0 {
		stop = trans - 1
	}
	return
}

func calcCoresPerMonad(cpuCores int,monads int) (cores int) {
	cores = cpuCores / monads

	if (cpuCores % monads) > 0 	{
		cores++
		}

	return cores
}