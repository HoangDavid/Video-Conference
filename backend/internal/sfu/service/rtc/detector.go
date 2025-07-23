package rtc

import (
	"context"
	"time"
	"vidcall/internal/sfu/domain"
)

type DetectorObj struct {
	*domain.DetectorObj
}

func NewDetector(ctx context.Context, interval time.Duration, margin int) domain.Detector {

	d := &DetectorObj{
		DetectorObj: &domain.DetectorObj{
			Sum:      make(map[string]int),
			Count:    make(map[string]int),
			Winner:   make(chan string, 1),
			FindFlag: make(chan struct{}, 1),
			Margin:   margin,
		},
	}

	go d.loop(ctx, interval)

	return d
}

func (d *DetectorObj) Sample(id string, lvl int) {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	if d.Current == "" {
		d.Current = id
		select {
		case d.Winner <- id:
		default:
		}
	}

	d.Sum[id] += lvl
	d.Count[id]++
}

func (d *DetectorObj) Remove(id string) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	delete(d.Sum, id)
	delete(d.Count, id)

	if d.Current == id {
		d.Current = ""
		select {
		case d.FindFlag <- struct{}{}:
		default:
		}
	}
}

func (d *DetectorObj) ActiveSpeaker() <-chan string {
	return d.Winner
}

func (d *DetectorObj) loop(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			close(d.Winner)
			return
		case <-t.C:
			d.findActiveSpeaker()
		case <-d.FindFlag:
			d.findActiveSpeaker()
		}
	}
}

func (d *DetectorObj) findActiveSpeaker() {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	if len(d.Sum) == 0 {
		return
	}

	best := d.Current
	bestAvg := 128
	curAvg := 128

	if c := d.Count[d.Current]; c > 0 {
		curAvg = d.Sum[d.Current] / c
	}

	for id, c := range d.Count {
		if c == 0 {
			continue
		}

		avg := d.Sum[id] / c
		if avg < bestAvg {
			bestAvg = avg
			best = id
		}
		d.Sum[id] = 0
		d.Count[id] = 0
	}

	//  if speaker left
	if d.Current == "" && best != "" {
		d.Current = best
		select {
		case d.Winner <- best:
		default:
		}
		return
	}

	// update the new active speaker
	if best != d.Current && bestAvg+d.Margin < curAvg {
		d.Current = best
		select {
		case d.Winner <- best:
		default:
		}
	}
}
