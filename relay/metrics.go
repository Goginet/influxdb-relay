package relay

import (
	"fmt"
	"net"
	"time"
)

type MetricsCollector struct {
	url      string
	name     string
	interval time.Duration
}

func (m MetricsCollector) RunMetricsCollector(list *bufferList) {
	for {
		list.cond.L.Lock()

		m.sendMetric(fmt.Sprintf("influxdb_relay,name=%s,backend=%s buffer_used=%d,buffer_size=%d", m.name, list.name, list.size, list.maxSize))

		list.cond.L.Unlock()
		time.Sleep(m.interval)
	}
}

func (m MetricsCollector) sendMetric(str string) {
	conn, err := net.Dial("udp", m.url)
	if err != nil {
		fmt.Printf("[E] Error! : %v", err)
		return
	}
	defer conn.Close()

	//simple write
	conn.Write([]byte(str))
}
