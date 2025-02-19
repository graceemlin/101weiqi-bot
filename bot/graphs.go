package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
	"log"
	"os"
	"strconv"
)

type fill_type int

const (
	standard fill_type = iota
	winner
	loser
)

var HISTOGRAMFILE string

func graph_create(stat_array *[23]Statistic, user string, curr_fill fill_type) {
	// extract relevant values from the array of Statistics
	completion_times := make(plotter.Values, 22)
	for pop := 1; pop <= 22; pop++ {
		level_time, level_time_error := strconv.Atoi(stat_array[pop].Time)
		if level_time_error != nil {
			continue
		}
		completion_times[pop-1] = float64(level_time)
	}

	// construct histogram from completion time data
	histogram := plot.New()
	bars, bar_creation_error := plotter.NewBarChart(completion_times, 10)
	if bar_creation_error != nil {
		log.Fatal(bar_creation_error)
	}

	// bar color handling
	switch curr_fill {
	case standard:
		bars.Color = color.RGBA{B: 128, A: 255}
	case winner:
		bars.Color = color.RGBA{G: 128, A: 255}
	case loser:
		bars.Color = color.RGBA{R: 128, A: 255}
	}

	histogram.Add(bars)

	// label histogram
	histogram.Title.Text = "Skill Test Histogram: " + user
	histogram.X.Label.Text = "Skill Test Level"
	histogram.Y.Label.Text = "Completion Time"

	// add title padding
	histogram.Title.Padding = vg.Points(30)

	// label the x-axis
	histogram.X.Tick.Marker = plot.TickerFunc(func(min float64, max float64) []plot.Tick {
		x_ticks := make([]plot.Tick, 22)
		for pop := 1; pop <= 22; pop++ {
			label := fmt.Sprintf("%s", pop_to_level(pop))
			x_ticks[pop-1] = plot.Tick{Value: float64(pop - 1), Label: label}
		}

		return x_ticks
	})

	histogram.Y.Tick.Marker = plot.TickerFunc(func(min float64, max float64) []plot.Tick {
		var y_ticks []plot.Tick
		for seconds := min; seconds <= max; seconds += 45 {
			y_ticks = append(y_ticks, plot.Tick{Value: float64(seconds), Label: fmt.Sprintf("%ds", int(seconds))})
		}

		return y_ticks
	})

	// rotate x-axis labels
	histogram.X.Tick.Label.Rotation = -45

	// save the plot to a file
	save_err := histogram.Save(6*vg.Inch, 4*vg.Inch, HISTOGRAMFILE)
	if save_err != nil {
		log.Fatal(save_err)
	}
}

func graph_print(user string, message *discordgo.MessageCreate, session *discordgo.Session) {
	// open histogram image
	graph_file, graph_file_open_error := os.Open(HISTOGRAMFILE)
	if graph_file_open_error != nil {
		log.Fatal("Error opening histogram image:", graph_file_open_error)
	}
	defer graph_file.Close()

	// send the image to the output channel
	_, print_error := session.ChannelFileSend(message.ChannelID, HISTOGRAMFILE, graph_file)
	if print_error != nil {
		log.Fatal("Error sending histogram image:", print_error)
	}
}
