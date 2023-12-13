package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluenviron/mediamtx/internal/logger"
)

type recFormatMPEGTSSegment struct {
	f         *recFormatMPEGTS
	startDTS  time.Duration
	lastFlush time.Duration

	created time.Time
	fpath   string
	fi      *os.File
}

func newRecFormatMPEGTSSegment(f *recFormatMPEGTS, startDTS time.Duration) *recFormatMPEGTSSegment {
	s := &recFormatMPEGTSSegment{
		f:         f,
		startDTS:  startDTS,
		lastFlush: startDTS,
		created:   timeNow(),
	}

	f.dw.setTarget(s)

	return s
}

func (s *recFormatMPEGTSSegment) close() error {
	err := s.f.bw.Flush()

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
							paths[len(paths)-1],
						))
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

func (s *recFormatMPEGTSSegment) Write(p []byte) (int, error) {
	if s.fi == nil {
		s.fpath = encodeRecordPath(&recordPathParams{time: s.created}, s.f.a.resolvedPath)
		s.f.a.wrapper.Log(logger.Debug, "creating segment %s", s.fpath)

		err := os.MkdirAll(filepath.Dir(s.fpath), 0o755)
		if err != nil {
			return 0, err
		}

		fi, err := os.Create(s.fpath)
		if err != nil {
			return 0, err
		}

		if s.f.a.stor.Use {
			paths := strings.Split(s.fpath, "/")
			pathRec := strings.Join(paths[:len(paths)-1], "/")
			err := s.f.a.stor.Req.ExecQuery(
				fmt.Sprintf(
					s.f.a.stor.Sql.InsertPath,
					s.f.a.wrapper.PathName,
					pathRec+"/",
					paths[len(paths)-1],
					time.Now().Format("2006-01-02 15:04:05"),
				),
			)
			if err != nil {
				os.Remove(s.fpath)
				return 0, err
			}
		}

		s.f.a.wrapper.OnSegmentCreate(s.fpath)

		s.fi = fi
	}

	return s.fi.Write(p)
}
