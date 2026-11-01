//go:build ignore

package main

import (
	"encoding/json/v2"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type cue struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

func main() {
	input := flag.String("input", "tmp/blog-tutorial.mp4", "original video")
	output := flag.String("output", "tmp/blog-tutorial-captioned.mp4", "captioned video")
	source := flag.String("cues", "tutorial/captions.json", "caption timing and text")
	work := flag.String("work", "tmp/blog-captions/generated", "intermediate files")
	font := flag.String("font", "/System/Library/Fonts/Helvetica.ttc", "ImageMagick font name or path")
	flag.Parse()
	if err := caption(*input, *output, *source, *work, *font); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func caption(input, output, source, work, font string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	var cues []cue
	if err := json.Unmarshal(data, &cues); err != nil {
		return fmt.Errorf("read captions: %w", err)
	}
	probe, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=nw=1:nk=1", input).Output()
	if err != nil {
		return fmt.Errorf("read video duration: %w", err)
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(probe)), 64)
	if err != nil {
		return err
	}
	previous := 0.0
	for i, cue := range cues {
		if cue.Start < previous || cue.End <= cue.Start || cue.End > duration || strings.TrimSpace(cue.Text) == "" {
			return fmt.Errorf("invalid caption %d: times must be ordered, non-overlapping, and within the video", i+1)
		}
		previous = cue.End
	}
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	var srt, vtt, playlist strings.Builder
	vtt.WriteString("WEBVTT\n\n")
	playlist.WriteString("ffconcat version 1.0\n")
	previous = 0
	for i, cue := range cues {
		fmt.Fprintf(&srt, "%d\n%s --> %s\n%s\n\n", i+1, timestamp(cue.Start, ","), timestamp(cue.End, ","), cue.Text)
		fmt.Fprintf(&vtt, "%s --> %s\n%s\n\n", timestamp(cue.Start, "."), timestamp(cue.End, "."), cue.Text)
		if cue.Start > previous {
			fmt.Fprintf(&playlist, "file 'blank.png'\nduration %.3f\n", cue.Start-previous)
		}
		name := fmt.Sprintf("%03d.png", i+1)
		if err := band(filepath.Join(work, name), cue.Text, font); err != nil {
			return err
		}
		fmt.Fprintf(&playlist, "file '%s'\nduration %.3f\n", name, cue.End-cue.Start)
		previous = cue.End
	}
	if duration > previous {
		fmt.Fprintf(&playlist, "file 'blank.png'\nduration %.3f\n", duration-previous)
	}
	playlist.WriteString("file 'blank.png'\n")
	if err := band(filepath.Join(work, "blank.png"), "", font); err != nil {
		return err
	}
	base := strings.TrimSuffix(output, filepath.Ext(output))
	for path, content := range map[string]string{
		base + ".srt":                         srt.String(),
		base + ".vtt":                         vtt.String(),
		filepath.Join(work, "bands.ffconcat"): playlist.String(),
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return run("ffmpeg", "-y", "-loglevel", "warning",
		"-i", input, "-f", "concat", "-safe", "0", "-i", filepath.Join(work, "bands.ffconcat"),
		"-filter_complex", "[0:v][1:v]vstack=inputs=2:shortest=1[video]", "-map", "[video]",
		"-t", fmt.Sprintf("%.3f", duration), "-r", "15", "-c:v", "libx264", "-preset", "fast",
		"-crf", "18", "-pix_fmt", "yuv420p", "-movflags", "+faststart", output)
}

func timestamp(seconds float64, separator string) string {
	ms := int64(math.Round(seconds * 1000))
	return fmt.Sprintf("%02d:%02d:%02d%s%03d", ms/3600000, ms/60000%60, ms/1000%60, separator, ms%1000)
}

func band(path, text, font string) error {
	args := []string{"-size", "1440x104", "xc:#171b26"}
	if text != "" {
		args = append(args, "-font", font, "-pointsize", "30", "-fill", "white", "-gravity", "center", "-annotate", "+0+0", text)
	}
	return run("magick", append(args, "PNG24:"+path)...)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
