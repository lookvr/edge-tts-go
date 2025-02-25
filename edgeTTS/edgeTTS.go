package edgeTTS

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"golang.org/x/crypto/ssh/terminal"
)

type EdgeTTS struct {
	communicator *Communicate
	tasks        []*CommunicateTextTask
	outCome      io.WriteCloser
}

type Args struct {
	Text           string
	Voice          string
	Proxy          string
	Rate           string
	Volume         string
	WordsInCue     float64
	WriteMedia     string
	WriteSubtitles string
}

func isTerminal(file *os.File) bool {
	return terminal.IsTerminal(int(file.Fd()))
}

func PrintVoices(locale string) {
	// Print all available voices.
	voices, err := listVoices()
	if err != nil {
		log.Fatalf("Failed to listVoices: %v\n", err)
		return
	}
	sort.Slice(voices, func(i, j int) bool {
		return voices[i].ShortName < voices[j].ShortName
	})

	// log.Printf("error: %+v \n", voices)
	filterFieldName := map[string]bool{
		"SuggestedCodec": true,
		"FriendlyName":   true,
		"Status":         true,
		"VoiceTag":       true,
		"Language":       true,
	}

	for _, voice := range voices {
		lenLocale := len(locale)
		if lenLocale > 0 {
			tempLocale := voice.Locale
			if len(voice.Locale) > lenLocale {
				tempLocale = voice.Locale[:lenLocale]
			}

			if tempLocale != locale {
				continue
			}
		}

		fmt.Printf("\n")
		t := reflect.TypeOf(voice)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldName := field.Name
			if filterFieldName[fieldName] {
				continue
			}
			fieldValue := reflect.ValueOf(voice).Field(i).Interface()
			fmt.Printf("%s: %v\n", fieldName, fieldValue)
		}
	}
}

func NewTTS(args Args) *EdgeTTS {
	if isTerminal(os.Stdin) && isTerminal(os.Stdout) && args.WriteMedia == "" {
		fmt.Fprintln(os.Stderr, "Warning: TTS output will be written to the terminal. Use --write-media to write to a file.")
		fmt.Fprintln(os.Stderr, "Press Ctrl+C to cancel the operation. Press Enter to continue.")
		fmt.Scanln()
	}
	if _, err := os.Stat(args.WriteMedia); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(args.WriteMedia), 0755)
		if err != nil {
			log.Fatalf("Failed to create dir: %v\n", err)
			return nil
		}
	}

	file, err := os.OpenFile(args.WriteMedia, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file: %v\n", err)
		return nil
	}

	tts := NewCommunicate().WithVoice(args.Voice).WithRate(args.Rate).WithVolume(args.Volume)
	tts.openWs()
	return &EdgeTTS{
		communicator: tts,
		outCome:      file,
		tasks:        []*CommunicateTextTask{},
	}
}

func (eTTS *EdgeTTS) task(text string, voice string, rate string, volume string) *CommunicateTextTask {
	return &CommunicateTextTask{
		text: text,
		option: CommunicateTextOption{
			voice:  voice,
			rate:   rate,
			volume: volume,
		},
	}
}

func (eTTS *EdgeTTS) AddTextDefault(text string) *EdgeTTS {
	eTTS.tasks = append(eTTS.tasks, eTTS.task(text, "", "", ""))
	return eTTS
}

func (eTTS *EdgeTTS) AddTextWithVoice(text string, voice string) *EdgeTTS {
	eTTS.tasks = append(eTTS.tasks, eTTS.task(text, voice, "", ""))
	return eTTS
}

func (eTTS *EdgeTTS) AddText(text string, voice string, rate string, volume string) *EdgeTTS {
	eTTS.tasks = append(eTTS.tasks, eTTS.task(text, voice, rate, volume))
	return eTTS
}

func (eTTS *EdgeTTS) Speak() {
	defer eTTS.communicator.close()
	defer eTTS.outCome.Close()

	go eTTS.communicator.allocateTask(eTTS.tasks)
	eTTS.communicator.createPool()
	for _, task := range eTTS.tasks {
		log.Printf("eTTS.tasks.id: %d \n", task.id)
		n, err := eTTS.outCome.Write(task.speechData)
		if err != nil {
			log.Fatalf("outCome.Write.error: %v", err)
			return
		}

		log.Printf("outCome.Write.bytes: %d \n", n)
	}
}
