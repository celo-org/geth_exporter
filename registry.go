package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

const namespace = "geth"

type registry struct {
	c *collector
	*prometheus.Registry
}

func newRegistry(ipcPath string, rawFilters []string) *registry {
	return &registry{
		c:        newCollector(ipcPath, rawFilters),
		Registry: prometheus.NewRegistry(),
	}
}

// Gather implements the prometheus.Registerer interface
func (r *registry) Gather() (families []*dto.MetricFamily, err error) {
	fs, err := r.Registry.Gather()
	if err != nil {
		log.Println(err)
	} else {
		families = fs
	}

	fm, err := r.c.collect()
	if err != nil {
		log.Printf("Error collecting: %+v", err)
		return
	}

	for k, v := range fm {
		mf, err := r.buildMetricFamily(k, v)
		if err != nil {
			log.Printf("error building metric: %v", err)
			continue
		}

		families = append(families, mf)

	}

	return
}

func (r *registry) buildMetricFamily(name string, stringValue string) (*dto.MetricFamily, error) {
	value, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return nil, err
	}

	metricFamily := &dto.MetricFamily{}
	m := &dto.Metric{}
	m.Label = make([]*dto.LabelPair, 0)

	if strings.HasSuffix(name, "_overall") {
		m.Counter = &dto.Counter{Value: proto.Float64(value)}
		metricFamily.Type = dto.MetricType_COUNTER.Enum()
	} else {
		m.Gauge = &dto.Gauge{Value: proto.Float64(value)}
		metricFamily.Type = dto.MetricType_GAUGE.Enum()
	}

	metricFamily.Name = proto.String(fmt.Sprintf("%s_%s", namespace, name))
	metricFamily.Help = proto.String("")
	metricFamily.Metric = []*dto.Metric{m}

	return metricFamily, nil
}
