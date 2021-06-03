package main

import (
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fanaticscripter/EggContractor/api"
	"github.com/fanaticscripter/EggContractor/util"

	"github.com/fanaticscripter/Egg/wasmegg/_common/eiafx"
)

type ShipParameters = api.ArtifactsConfigurationResponse_MissionParameters

type ship struct {
	*ShipParameters
	Sensors                    string
	LaunchesToAdvance          uint32
	TimeToAdvanceStd           time.Duration
	TimeToAdvancePro           time.Duration
	CumulativeTimeToAdvanceStd time.Duration
	CumulativeTimeToAdvancePro time.Duration
}

type fuel struct {
	Egg    api.EggType
	Amount float64
}

func main() {
	err := eiafx.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	config := eiafx.Config

	var ships []*ship
	var cumulativeTimeToAdvance time.Duration
	stdConcurrency := time.Duration(1)
	proConcurrency := time.Duration(3)
	for _, s := range config.MissionParameters {
		launchesToAdvance := shipRequiredLaunchesToAdvance(s.Ship)
		var timeToAdvance time.Duration
		if s.Ship == api.MissionInfo_CHICKEN_ONE {
			// Forget about the nuance of 2 tutorial missions + 2 short missions
			timeToAdvance = 3 * 20 * time.Minute
			cumulativeTimeToAdvance = timeToAdvance
		} else if s.Ship == api.MissionInfo_HENERPRISE {
			// Do not display cumulative for Henerprise
			cumulativeTimeToAdvance = 0
		} else {
			for _, t := range s.Durations {
				if t.DurationType == api.MissionInfo_SHORT {
					timeToAdvance = time.Duration(launchesToAdvance) * util.DoubleToDuration(t.Seconds)
					break
				}
			}
			if timeToAdvance == 0 {
				panic(fmt.Sprintf("short mission not found for ship %s", s.Ship))
			}
			cumulativeTimeToAdvance += timeToAdvance
		}
		ships = append(ships, &ship{
			ShipParameters:             s,
			Sensors:                    shipSensors(s.Ship),
			LaunchesToAdvance:          launchesToAdvance,
			TimeToAdvanceStd:           timeToAdvance / stdConcurrency,
			TimeToAdvancePro:           timeToAdvance / proConcurrency,
			CumulativeTimeToAdvanceStd: cumulativeTimeToAdvance / stdConcurrency,
			CumulativeTimeToAdvancePro: cumulativeTimeToAdvance / proConcurrency,
		})
	}

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"eggiconpath":      eggIconPath,
		"fmtduration":      util.FormatDurationWhole,
		"fuels":            missionFuels,
		"iconurl":          iconURL,
		"numfmt":           util.NumfmtWhole,
		"shipiconpath":     shipIconPath,
		"seconds2duration": util.DoubleToDuration,
	}).ParseGlob("templates/*/*.html"))
	err = os.MkdirAll("src", 0o755)
	if err != nil {
		log.Fatalf("mkdir -p src failed: %s", err)
	}
	output, err := os.Create("src/index.html")
	if err != nil {
		log.Fatalf("failed to open src/index.html for writing: %s", err)
	}
	defer output.Close()
	err = tmpl.ExecuteTemplate(output, "index.html", struct {
		Ships []*ship
	}{
		Ships: ships,
	})
	if err != nil {
		log.Fatalf("failed to render template: %s", err)
	}
}

func iconURL(relpath string, size int) string {
	dir := strconv.Itoa(size)
	if size <= 0 {
		dir = "orig"
	}
	return fmt.Sprintf("https://eggincassets.tcl.sh/%s/%s", dir, relpath)
}

func shipIconPath(ship api.MissionInfo_Spaceship) string {
	return "egginc/" + ship.IconFilename()
}

func eggIconPath(egg api.EggType) string {
	return "egginc/" + egg.IconFilename()
}

func shipSensors(ship api.MissionInfo_Spaceship) string {
	switch ship {
	case api.MissionInfo_CHICKEN_ONE:
		fallthrough
	case api.MissionInfo_CHICKEN_NINE:
		return "Basic"
	case api.MissionInfo_CHICKEN_HEAVY:
		fallthrough
	case api.MissionInfo_BCR:
		return "Intermediate"
	case api.MissionInfo_MILLENIUM_CHICKEN:
		fallthrough
	case api.MissionInfo_CORELLIHEN_CORVETTE:
		fallthrough
	case api.MissionInfo_GALEGGTICA:
		return "Advanced"
	case api.MissionInfo_CHICKFIANT:
		fallthrough
	case api.MissionInfo_VOYEGGER:
		return "Cutting Edge"
	case api.MissionInfo_HENERPRISE:
		return "Next Generation"
	}
	return ""
}

func shipRequiredLaunchesToAdvance(ship api.MissionInfo_Spaceship) uint32 {
	switch ship {
	case api.MissionInfo_CHICKEN_ONE:
		return 4
	case api.MissionInfo_CHICKEN_NINE:
		return 6
	case api.MissionInfo_CHICKEN_HEAVY:
		return 12
	case api.MissionInfo_BCR:
		return 15
	case api.MissionInfo_MILLENIUM_CHICKEN:
		return 18
	case api.MissionInfo_CORELLIHEN_CORVETTE:
		return 21
	case api.MissionInfo_GALEGGTICA:
		return 24
	case api.MissionInfo_CHICKFIANT:
		return 27
	case api.MissionInfo_VOYEGGER:
		return 30
	}
	return 0
}

func missionFuels(ship api.MissionInfo_Spaceship, durationType api.MissionInfo_DurationType) []fuel {
	return _fuels[ship][durationType]
}

// This shit is typed by hand.
var _fuels = map[api.MissionInfo_Spaceship]map[api.MissionInfo_DurationType][]fuel{
	api.MissionInfo_CHICKEN_ONE: {
		api.MissionInfo_TUTORIAL: {
			{api.EggType_ROCKET_FUEL, 1e5},
		},
		api.MissionInfo_SHORT: {
			{api.EggType_ROCKET_FUEL, 2e6},
		},
		api.MissionInfo_LONG: {
			{api.EggType_ROCKET_FUEL, 3e6},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_ROCKET_FUEL, 10e6},
		},
	},
	api.MissionInfo_CHICKEN_NINE: {
		api.MissionInfo_SHORT: {
			{api.EggType_ROCKET_FUEL, 10e6},
		},
		api.MissionInfo_LONG: {
			{api.EggType_ROCKET_FUEL, 15e6},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_ROCKET_FUEL, 25e6},
		},
	},
	api.MissionInfo_CHICKEN_HEAVY: {
		api.MissionInfo_SHORT: {
			{api.EggType_ROCKET_FUEL, 100e6},
		},
		api.MissionInfo_LONG: {
			{api.EggType_ROCKET_FUEL, 50e6},
			{api.EggType_FUSION, 5e6},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_ROCKET_FUEL, 75e6},
			{api.EggType_FUSION, 25e6},
		},
	},
	api.MissionInfo_BCR: {
		api.MissionInfo_SHORT: {
			{api.EggType_ROCKET_FUEL, 250e6},
			{api.EggType_FUSION, 50e6},
		},
		api.MissionInfo_LONG: {
			{api.EggType_ROCKET_FUEL, 400e6},
			{api.EggType_FUSION, 75e6},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_SUPERFOOD, 5e6},
			{api.EggType_ROCKET_FUEL, 300e6},
			{api.EggType_FUSION, 100e6},
		},
	},
	api.MissionInfo_MILLENIUM_CHICKEN: {
		api.MissionInfo_SHORT: {
			{api.EggType_FUSION, 5e9},
			{api.EggType_GRAVITON, 1e9},
		},
		api.MissionInfo_LONG: {
			{api.EggType_FUSION, 7e9},
			{api.EggType_GRAVITON, 5e9},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_SUPERFOOD, 10e6},
			{api.EggType_FUSION, 10e9},
			{api.EggType_GRAVITON, 15e9},
		},
	},
	api.MissionInfo_CORELLIHEN_CORVETTE: {
		api.MissionInfo_SHORT: {
			{api.EggType_FUSION, 15e9},
			{api.EggType_GRAVITON, 2e9},
		},
		api.MissionInfo_LONG: {
			{api.EggType_FUSION, 20e9},
			{api.EggType_GRAVITON, 3e9},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_SUPERFOOD, 500e6},
			{api.EggType_FUSION, 25e9},
			{api.EggType_GRAVITON, 5e9},
		},
	},
	api.MissionInfo_GALEGGTICA: {
		api.MissionInfo_SHORT: {
			{api.EggType_FUSION, 50e9},
			{api.EggType_GRAVITON, 10e9},
		},
		api.MissionInfo_LONG: {
			{api.EggType_FUSION, 75e9},
			{api.EggType_GRAVITON, 25e9},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_FUSION, 100e9},
			{api.EggType_GRAVITON, 50e9},
			{api.EggType_ANTIMATTER, 1e9},
		},
	},
	api.MissionInfo_CHICKFIANT: {
		api.MissionInfo_SHORT: {
			{api.EggType_DILITHIUM, 200e9},
			{api.EggType_ANTIMATTER, 50e9},
		},
		api.MissionInfo_LONG: {
			{api.EggType_DILITHIUM, 250e9},
			{api.EggType_ANTIMATTER, 150e9},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_TACHYON, 25e9},
			{api.EggType_DILITHIUM, 250e9},
			{api.EggType_ANTIMATTER, 250e9},
		},
	},
	api.MissionInfo_VOYEGGER: {
		api.MissionInfo_SHORT: {
			{api.EggType_DILITHIUM, 1e12},
			{api.EggType_ANTIMATTER, 1e12},
		},
		api.MissionInfo_LONG: {
			{api.EggType_DILITHIUM, 1.5e12},
			{api.EggType_ANTIMATTER, 1.5e12},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_TACHYON, 100e9},
			{api.EggType_DILITHIUM, 2e12},
			{api.EggType_ANTIMATTER, 2e12},
		},
	},
	api.MissionInfo_HENERPRISE: {
		api.MissionInfo_SHORT: {
			{api.EggType_DILITHIUM, 2e12},
			{api.EggType_ANTIMATTER, 2e12},
		},
		api.MissionInfo_LONG: {
			{api.EggType_DILITHIUM, 3e12},
			{api.EggType_ANTIMATTER, 3e12},
			{api.EggType_DARK_MATTER, 3e12},
		},
		api.MissionInfo_EPIC: {
			{api.EggType_TACHYON, 1e12},
			{api.EggType_DILITHIUM, 3e12},
			{api.EggType_ANTIMATTER, 3e12},
			{api.EggType_DARK_MATTER, 3e12},
		},
	},
}
