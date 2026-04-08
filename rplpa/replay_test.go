package rplpa

import (
	"os"
	"testing"
)

func TestParseStableReplay(t *testing.T) {
	parseTest(t, "data/replay1.osr")
}

func TestParseLazerReplay(t *testing.T) {
	parseTest(t, "data/lazer.osr")
}

func TestWriteStableReplay(t *testing.T) {
	writeTest(t, "data/replay1.osr", "data/replay1_out.osr")
}

func TestWriteLazerReplay(t *testing.T) {
	writeTest(t, "data/lazer.osr", "data/lazer_out.osr")
}

func writeTest(t *testing.T, filename, out string) {
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Error("Could not read replay, Doesn't exist?")
	}

	p, err := ParseReplay(b)
	if err != nil {
		t.Error("Could not parse replay", err)
	}

	p.Username = "harvywtf"
	p.ScoreID = 0

	if p != nil {
		data, err := WriteReplay(p)
		if err != nil {
			t.Error("Could not write replay", err)
		}

		file, err := os.Create(out)
		if err != nil {
			t.Error("Could not create output file", err)
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			t.Error("Could not write to output file", err)
		}
	}
}

func parseTest(t *testing.T, filename string) {
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Error("Could not read replay, Doesn't exist?")
	}

	p, err := ParseReplay(b)
	if err != nil {
		t.Error("Could not parse replay", err)
	}

	if p != nil {
		t.Log("PlayMode: ", p.PlayMode)
		t.Log("OsuVersion: ", p.OsuVersion)
		t.Log("BeatmapMD5: ", p.BeatmapMD5)
		t.Log("Username: ", p.Username)
		t.Log("ReplayMD5: ", p.ReplayMD5)
		t.Log("Count300: ", p.Count300)
		t.Log("Count100: ", p.Count100)
		t.Log("Count50: ", p.Count50)
		t.Log("CountGeki: ", p.CountGeki)
		t.Log("CountKatu: ", p.CountKatu)
		t.Log("CountMiss: ", p.CountMiss)
		t.Log("Score: ", p.Score)
		t.Log("MaxCombo: ", p.MaxCombo)
		t.Log("Fullcombo: ", p.Fullcombo)
		t.Log("Mods: ", p.Mods)
		t.Log("LifebarGraph: ", p.LifebarGraph)
		t.Log("Timestamp: ", p.Timestamp)
		t.Log("ScoreID: ", p.ScoreID)
		t.Log("InputEvents", len(p.ReplayData))
		// if p.ScoreInfo != nil {
		// 	t.Log("ScoreInfo Mods: ", len(p.ScoreInfo.Mods))

		// 	if len(p.ScoreInfo.Mods) > 0 {
		// 		for _, m := range p.ScoreInfo.Mods {
		// 			t.Log("ScoreInfo Mod: ", *m)
		// 		}
		// 	}
		// 	if p.ScoreInfo.Statistics != nil {
		// 		t.Log("ScoreInfo Statistics: ", p.ScoreInfo.Statistics)
		// 	}

		// 	if p.ScoreInfo.MaximumStatistics != nil {
		// 		t.Log("ScoreInfo MaximumStatistics: ", p.ScoreInfo.MaximumStatistics)
		// 	}
		// } else {
		// 	t.Log("ScoreInfo is nil due to EOF (stable play)")
		// }
	}
}
