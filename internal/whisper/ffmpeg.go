package whisper

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ConvertOGGToWAV(inFile string, outFile string) error {
	return ffmpeg.Input(inFile).
		Output(outFile, ffmpeg.KwArgs{"ar": "16000"}).
		Run()
}
