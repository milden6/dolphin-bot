package handlers

import (
	"crypto/rand"
	"dolphin-bot/internal/whisper"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

const maxVoiceDuration = 60 // in seconds

const tempAudioDir = "static/temp_audio/"

var whisperAPI = whisper.NewAPIClient("http://127.0.0.1:8080")

func HandleVoiceMsg(ctx tele.Context) error {
	startHandling := time.Now()
	defer func() {
		slog.Info("Voice message handled", "username", ctx.Sender().Username, "request processed", time.Since(startHandling))
	}()

	voice := ctx.Message().Voice

	if voice.Duration > maxVoiceDuration {
		slog.Error("Max voice duration", "username", ctx.Sender().Username)

		return ctx.Send("Извини, но это слишком долго слушать =(", &tele.SendOptions{ReplyTo: ctx.Message()})
	}

	// send immediately
	err := ctx.Send("Так-так, дайка послушать 🧐")
	if err != nil {
		slog.Error("Failed to send msg to telegram", "error", err.Error())
	}

	go func() {
		sendOnErr := func() {
			err := ctx.Send("У меня что-то сломалось 😥\n Попробуй позже 🙏", &tele.SendOptions{ReplyTo: ctx.Message()})
			if err != nil {
				slog.Error("Failed to send msg to telegram", "error", err.Error())
			}
		}

		fileReader, err := ctx.Bot().File(&voice.File)
		if err != nil {
			slog.Error("Failed to get voice file from telegram server", "error", err.Error())
			sendOnErr()

			return
		}
		defer func() {
			err := fileReader.Close()
			if err != nil {
				slog.Error("Failed to close voice fileReader", "error", err.Error())
			}
		}()

		inFile, outFile, err := saveVoiceToDisk(fileReader)
		if err != nil {
			slog.Error("Failed to save voice file to disk", "error", err.Error())
			sendOnErr()

			return
		}

		err = whisper.ConvertOGGToWAV(inFile, outFile)
		if err != nil {
			slog.Error("Failed to convert ogg to wav", "error", err.Error())
			sendOnErr()

			return
		}

		err = os.Remove(inFile)
		if err != nil {
			slog.Error(err.Error())
		}

		decodedText, err := whisperAPI.DoInference(outFile)
		if err != nil {
			slog.Error("Failed on whisper inference", "error", err.Error())
			sendOnErr()

			return
		}

		// remove new line for message
		decodedText = strings.ReplaceAll(decodedText, "\n", "")
		err = ctx.Send(decodedText, &tele.SendOptions{ReplyTo: ctx.Message()})
		if err != nil {
			slog.Error("Failed to send msg to telegram", "error", err.Error())
		}

		err = os.Remove(outFile)
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	return nil
}

func saveVoiceToDisk(file io.Reader) (string, string, error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	uuid, err := uuidv7()
	if err != nil {
		return "", "", err
	}

	fileID := fmt.Sprintf("%x", uuid)
	tempInputFile := tempAudioDir + fileID + ".ogg"
	tempOutputFile := tempAudioDir + fileID + ".wav"

	err = os.WriteFile(tempInputFile, fileData, 0666)
	if err != nil {
		return "", "", err
	}

	return tempInputFile, tempOutputFile, nil
}

func uuidv7() ([16]byte, error) {
	// random bytes
	var value [16]byte
	_, err := rand.Read(value[:])
	if err != nil {
		return value, err
	}

	// current timestamp in ms
	timestamp := big.NewInt(time.Now().UnixMilli())

	// timestamp
	timestamp.FillBytes(value[0:6])

	// version and variant
	value[6] = (value[6] & 0x0F) | 0x70
	value[8] = (value[8] & 0x3F) | 0x80

	return value, nil
}
