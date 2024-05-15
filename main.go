package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

func (r *RingIntBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
	} else {
		r.pos++
		r.array[r.pos] = el
	}
}

func (r *RingIntBuffer) Get() []int {
	if r.pos <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos+1]
	r.pos = -1
	return output
}

func read(nextStage chan<- int, done chan bool) {
	log.Println("Запуск функции чтения из консоли")
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			log.Println("Введена команда exit, выход из приложения, закрытие канала done")
			fmt.Println("Программа завершена. ")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			log.Println("Получено не целое число: ", data)
			fmt.Println("Программа обрабатывает только целые числа")
			continue
		}
		log.Println("Получено целое число: ", i)
		nextStage <- i
	}
}

func negativeFilterStageInt(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {
	log.Println("Запущена функция фильтрации отрицательных чисел.")
	for {
		select {
		case data := <-previosStageChannel:
			log.Println("Функция фильтрации отрицательных чисел, получено из канала: ", data)
			if data > 0 {
				log.Println("Функция фильтрации отрицательных чисел пройдена, положительное число отправлено в следующий канал: ", data)
				nextStageChannel <- data
			}
		case <-done:
			log.Println("Канал done закрыт, выход из функции фильтрации отрицательных чисел")
			return
		}
	}
}

func notDividedThreeFunc(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {
	log.Println("Функция фильтрации деления на 3 запущена")
	for {
		select {
		case data := <-previosStageChannel:
			log.Println("Функция фильтрации деления на 3, получено из канала: ", data)
			if data%3 == 0 {
				log.Println("Функция фильтрации деления на 3 пройдена, число отправлено в следующий канал: ", data)
				nextStageChannel <- data
			}
		case <-done:
			log.Println("Канал done закрыт, выход из функции фильтрации деления на 3")
			return
		}
	}
}

func bufferStageFunc(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingIntBuffer(size)
	for {
		select {
		case data := <-previosStageChannel:
			log.Println("Запись значения в кольцевой буфер:", data)
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			if bufferData != nil {
				log.Println("Запись в канал из кольцового буфера по истечении", interval)
				for _, data := range bufferData {
					nextStageChannel <- data
				}
			}
		case <-done:
			log.Println("Канал done закрыт, выход из функции, для работы с кольцевым буфером")
			return
		}
	}
}
func main() {
	log.Println("Запуск приложения")
	input := make(chan int)
	done := make(chan bool)
	log.Println("Создание горутины для чтения из консоли")
	go read(input, done)

	negativeFilterChannel := make(chan int)
	log.Println("Создание горутины для фильтрации отрицательных чисел")
	go negativeFilterStageInt(input, negativeFilterChannel, done)

	notDividedThreeChannel := make(chan int)
	log.Println("Создание горутины для фильтрации деления на 3")
	go notDividedThreeFunc(negativeFilterChannel, notDividedThreeChannel, done)

	bufferedIntChannel := make(chan int)
	bufferSize := 10
	bufferDrainInterval := time.Second * 30
	log.Println("Создание горутины для работы кольцевого буфера")
	go bufferStageFunc(notDividedThreeChannel, bufferedIntChannel, done, bufferSize, bufferDrainInterval)

	for {
		select {
		case data := <-bufferedIntChannel:
			log.Println("Получение из канала кольцвого буфера обработанных данных:", data)
			fmt.Println("Обработанные данные: ", data)
		case <-done:
			log.Println("Канал done закрыт, завершение программы")
			return
		}
	}

}
