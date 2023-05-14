package formatprocessor

import (
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v3/pkg/formats"
	"github.com/pion/rtp"

	"github.com/aler9/mediamtx/internal/logger"
)

// UnitGeneric is a generic data unit.
type UnitGeneric struct {
	RTPPackets []*rtp.Packet
	NTP        time.Time
}

// GetRTPPackets implements Unit.
func (d *UnitGeneric) GetRTPPackets() []*rtp.Packet {
	return d.RTPPackets
}

// GetNTP implements Unit.
func (d *UnitGeneric) GetNTP() time.Time {
	return d.NTP
}

type formatProcessorGeneric struct {
	udpMaxPayloadSize int
}

func newGeneric(
	udpMaxPayloadSize int,
	forma formats.Format,
	generateRTPPackets bool,
	log logger.Writer,
) (*formatProcessorGeneric, error) {
	if generateRTPPackets {
		return nil, fmt.Errorf("we don't know how to generate RTP packets of format %+v", forma)
	}

	return &formatProcessorGeneric{
		udpMaxPayloadSize: udpMaxPayloadSize,
	}, nil
}

func (t *formatProcessorGeneric) Process(unit Unit, hasNonRTSPReaders bool) error {
	tunit := unit.(*UnitGeneric)

	pkt := tunit.RTPPackets[0]

	// remove padding
	pkt.Header.Padding = false
	pkt.PaddingSize = 0

	if pkt.MarshalSize() > t.udpMaxPayloadSize {
		return fmt.Errorf("payload size (%d) is greater than maximum allowed (%d)",
			pkt.MarshalSize(), t.udpMaxPayloadSize)
	}

	return nil
}

func (t *formatProcessorGeneric) UnitForRTPPacket(pkt *rtp.Packet, ntp time.Time) Unit {
	return &UnitGeneric{
		RTPPackets: []*rtp.Packet{pkt},
		NTP:        ntp,
	}
}
