package record

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aler9/writerseeker"
	"github.com/bluenviron/mediacommon/pkg/formats/fmp4"

	"github.com/bluenviron/mediamtx/internal/logger"
)

var timeNow = time.Now

func writeInit(f io.Writer, tracks []*recFormatFMP4Track) error {
	fmp4Tracks := make([]*fmp4.InitTrack, len(tracks))
	for i, track := range tracks {
		fmp4Tracks[i] = track.initTrack
	}

	init := fmp4.Init{
		Tracks: fmp4Tracks,
	}

	var ws writerseeker.WriterSeeker
	err := init.Marshal(&ws)
	if err != nil {
		return err
	}

	_, err = f.Write(ws.Bytes())
	return err
}

type recFormatFMP4Segment struct {
	f        *recFormatFMP4
	startDTS time.Duration

	fpath    string
	pathName string
	fi       *os.File
	curPart  *recFormatFMP4Part
}

func newRecFormatFMP4Segment(
	f *recFormatFMP4,
	startDTS time.Duration,
	pathName string,
) *recFormatFMP4Segment {
	return &recFormatFMP4Segment{
		f:        f,
		startDTS: startDTS,
		pathName: pathName,
	}
}

func (s *recFormatFMP4Segment) close() error {
	var err error

	if s.curPart != nil {
		err = s.curPart.close()
	}

	if s.fi != nil {
		s.f.a.wrapper.Log(logger.Debug, "closing segment %s", s.fpath)

		err2 := s.fi.Close()
		if err == nil {
			err = err2
		}
		if err2 == nil {
			s.f.a.wrapper.OnSegmentComplete(s.fpath)

			if s.f.a.stor.Use {
				stat, err3 := os.Stat(s.fpath)
				if err3 == nil {
					paths := strings.Split(s.fpath, "/")
					err4 := s.f.a.stor.Req.ExecQuery(
						fmt.Sprintf(
							s.f.a.stor.Sql.UpdateSize,
							fmt.Sprint(stat.Size()),
							time.Now().Format("2006-01-02 15:04:05"),
							paths[len(paths)-1]),
					)
					if err4 != nil {
						return err4
					}

					return err
				}
				err = err3
			}

		}
	}

	return err
}

func (s *recFormatFMP4Segment) record(track *recFormatFMP4Track, sample *sample) error {

	if s.curPart == nil {

		s.curPart = newRecFormatFMP4Part(s,
			s.f.nextSequenceNumber,
			sample.dts,
			s.pathName,
		)

		s.f.nextSequenceNumber++
	} else if s.curPart.duration() >= s.f.a.wrapper.PartDuration {
		err := s.curPart.close()
		s.curPart = nil
		if err != nil {
			return err
		}

		s.curPart = newRecFormatFMP4Part(s,
			s.f.nextSequenceNumber,
			sample.dts,
			s.pathName,
		)
		s.f.nextSequenceNumber++

	}

	return s.curPart.record(track, sample)
}
