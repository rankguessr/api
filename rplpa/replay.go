package rplpa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/itchio/lzma"
)

// NewReplay returns an Empty Replay
func NewReplay() *Replay {
	return &Replay{}
}

// ParseReplay parses a Replay and returns a *Replay
func ParseReplay(file []byte) (r *Replay, err error) {

	b := bytes.NewBuffer(file)
	r = NewReplay()

	if r.PlayMode, err = rInt8(b); err != nil {
		return nil, fmt.Errorf("reading PlayMode: %s", err)
	}

	if r.OsuVersion, err = rInt32(b); err != nil {
		return nil, fmt.Errorf("reading OsuVersion: %s", err)
	}

	if r.BeatmapMD5, err = rBString(b); err != nil {
		return nil, fmt.Errorf("reading BeatmapMD5: %s", err)
	}

	if r.Username, err = rBString(b); err != nil {
		return nil, fmt.Errorf("reading Username: %s", err)
	}

	if r.ReplayMD5, err = rBString(b); err != nil {
		return nil, fmt.Errorf("reading ReplayMD5: %s", err)
	}

	if r.Count300, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading Count300: %s", err)
	}

	if r.Count100, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading Count100: %s", err)
	}

	if r.Count50, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading Count50: %s", err)
	}

	if r.CountGeki, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading CountGeki: %s", err)
	}

	if r.CountKatu, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading CountKatu: %s", err)
	}

	if r.CountMiss, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading CountMiss: %s", err)
	}

	if r.Score, err = rInt32(b); err != nil {
		return nil, fmt.Errorf("reading Score: %s", err)
	}

	if r.MaxCombo, err = rUInt16(b); err != nil {
		return nil, fmt.Errorf("reading MaxCombo: %s", err)
	}

	if r.Fullcombo, err = rBool(b); err != nil {
		return nil, fmt.Errorf("reading Fullcombo: %s", err)
	}

	if r.Mods, err = rUInt32(b); err != nil {
		return nil, fmt.Errorf("reading Mods: %s", err)
	}

	var LifeBarRaw string
	if LifeBarRaw, err = rBString(b); err != nil {
		return nil, fmt.Errorf("reading LifeBar: %s", err)
	}

	r.LifebarGraph = parseLifebar(LifeBarRaw)

	var ts int64
	if ts, err = rInt64(b); err != nil {
		return nil, fmt.Errorf("reading Timestamp: %s", err)
	}

	r.Timestamp = timeFromTicks(ts)

	var cLength int32
	if cLength, err = rInt32(b); err != nil {
		return nil, fmt.Errorf("reading ReplayData length: %s", err)
	}

	if cLength > 0 {
		if r.ReplayData, err = rSlice(b, cLength); err != nil {
			return nil, fmt.Errorf("reading ReplayData: %s", err)
		}
	}

	if b.Len() >= 8 {
		if r.ScoreID, err = rInt64(b); err != nil {
			return nil, fmt.Errorf("reading ScoreID: %s", err)
		}
	} else if b.Len() >= 4 {
		var sID int32
		if sID, err = rInt32(b); err != nil {
			return nil, fmt.Errorf("reading ScoreID: %s", err)
		}

		r.ScoreID = int64(sID)
	}

	if b.Len() < 4 {
		return r, nil
	}

	var dLength int32
	if dLength, err = rInt32(b); err != nil {
		return nil, fmt.Errorf("reading ScoreInfo length: %s", err)
	}

	if dLength > 0 {
		scoreInfoRaw, err := rSlice(b, dLength)
		if err != nil {
			return nil, fmt.Errorf("reading ScoreInfo: %s", err)
		}

		scoreInfo, err := ParseCompressedScoreInfo(scoreInfoRaw)
		if err != nil {
			return nil, fmt.Errorf("parsing ScoreInfo: %s", err)
		}
		r.ScoreInfo = &scoreInfo
	} else if dLength == 0 {
		// This is a secondary case where it is a stable play not a lazer play
		r.ScoreInfo = nil
	}

	return
}

func WriteReplay(r *Replay) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	buf.Write(wInt8(r.PlayMode))
	buf.Write(wInt32(r.OsuVersion))
	buf.Write(wbString(r.BeatmapMD5))
	buf.Write(wbString(r.Username))
	buf.Write(wbString(r.ReplayMD5))
	buf.Write(wUInt16(r.Count300))
	buf.Write(wUInt16(r.Count100))
	buf.Write(wUInt16(r.Count50))
	buf.Write(wUInt16(r.CountGeki))
	buf.Write(wUInt16(r.CountKatu))
	buf.Write(wUInt16(r.CountMiss))
	buf.Write(wInt32(r.Score))
	buf.Write(wUInt16(r.MaxCombo))
	buf.Write(wBool(r.Fullcombo))
	buf.Write(wUInt32(r.Mods))
	buf.Write(wbString(serializeLifebar(r.LifebarGraph)))
	buf.Write(wInt64(ticksFromTime(r.Timestamp)))

	if len(r.ReplayData) > 0 {
		buf.Write(wInt32(int32(len(r.ReplayData))))
		buf.Write(r.ReplayData)
	} else {
		buf.Write(wInt32(0))
	}

	buf.Write(wUInt64(uint64(r.ScoreID)))

	if r.ScoreInfo != nil {
		scoreInfoBytes, err := serializeScoreInfo(*r.ScoreInfo)
		if err != nil {
			return nil, fmt.Errorf("writing ScoreInfo: %s", err)
		}

		buf.Write(wInt32(int32(len(scoreInfoBytes))))
		buf.Write(scoreInfoBytes)
	}

	return buf.Bytes(), nil
}

// https://stackoverflow.com/questions/33144967/what-is-the-c-sharp-datetimeoffset-equivalent-in-go/33161703#33161703

func timeFromTicks(ticks int64) time.Time {
	base := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	return time.Unix(ticks/10000000+base, ticks%10000000).UTC()
}

func ticksFromTime(tim time.Time) int64 {
	base := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	return tim.UnixNano()/100 - base*10000000
}

func parseLifebar(s string) []LifeBarGraph {
	var o []LifeBarGraph
	s = strings.Trim(s, ",")
	life := strings.Split(s, ",")
	for i := 0; i < len(life); i++ {
		y := strings.Split(life[i], "|")
		if len(y) < 2 {
			continue
		}
		f, err := strconv.ParseFloat(y[1], 32)
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(y[0], 32)
		o = append(o, LifeBarGraph{Time: int32(v), HP: float32(f)})
	}
	return o
}

func serializeLifebar(l []LifeBarGraph) string {
	sB := &strings.Builder{}

	for i, b := range l {
		sB.WriteString(fmt.Sprintf("%d|%f", b.Time, b.HP))

		if i < len(l)-1 {
			sB.WriteRune(',')
		}
	}

	return sB.String()
}

// ParseCompressed parses a compressed replay, (ReplayData)
func ParseCompressed(file []byte) (d []*ReplayData, err error) {
	b := bytes.NewBuffer(file)
	r := lzma.NewReader(b)
	defer r.Close()

	var data []byte
	if data, err = io.ReadAll(r); err != nil {
		return nil, fmt.Errorf("decompressing: %s", err)
	}

	events := strings.Split(strings.Trim(string(data), ","), ",")

	for i := 0; i < len(events); i++ {
		spl := strings.Split(events[i], "|")
		if len(spl) < 4 {
			continue
		}

		var timeFloat float64
		var MouseX float64
		var MouseY float64
		var keys int

		if timeFloat, err = strconv.ParseFloat(spl[0], 64); err != nil {
			return nil, fmt.Errorf("parsing Time on event %d: %s", i, err)
		}

		if MouseX, err = strconv.ParseFloat(spl[1], 64); err != nil {
			return nil, fmt.Errorf("parsing MouseX on event %d: %s", i, err)
		}

		if MouseY, err = strconv.ParseFloat(spl[2], 64); err != nil {
			return nil, fmt.Errorf("parsing MouseY on event %d: %s", i, err)
		}

		if keys, err = strconv.Atoi(spl[3]); err != nil {
			return nil, fmt.Errorf("parsing Keys on event %d: %s", i, err)
		}

		d = append(d, &ReplayData{
			Time:   timeFloat,
			MouseX: MouseX,
			MouseY: MouseY,
			KeyPressed: &KeyPressed{
				LeftClick:  keys&LEFTCLICK > 0,
				RightClick: keys&RIGHTCLICK > 0,
				Key1:       keys&KEY1 > 0,
				Key2:       keys&KEY2 > 0,
				Smoke:      keys&SMOKE > 0,
			},
		})
	}

	return
}

// ParseCompressedScoreInfo parses compressed ScoreInfo, (ScoreInfo)
func ParseCompressedScoreInfo(file []byte) (ScoreInfo, error) {
	b := bytes.NewBuffer(file)
	r := lzma.NewReader(b)
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return ScoreInfo{}, fmt.Errorf("decompressing: %s", err)
	}

	var scoreInfo ScoreInfo
	if err := json.Unmarshal(data, &scoreInfo); err != nil {
		return ScoreInfo{}, fmt.Errorf("parsing JSON: %s", err)
	}

	return scoreInfo, nil
}

func serializeScoreInfo(scoreInfo ScoreInfo) ([]byte, error) {
	data, err := json.Marshal(scoreInfo)
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %s", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 512))
	w := lzma.NewWriter(buf)

	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("compressing: %s", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("closing compressor: %s", err)
	}

	return buf.Bytes(), nil
}
