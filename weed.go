package main

import (
	"errors"
	"github.com/ginuerzh/weedo"
	"strconv"
	"strings"
)

var (
	InvalidFid = errors.New("invalid fid")
)

type WeedAdapter struct {
	client  *weedo.Client
	volumes map[int]string
}

func NewAdapter() *WeedAdapter {
	w := &WeedAdapter{}
	w.volumes = make(map[int]string)
	w.client = weedo.NewClient(weedUrl)
	return w
}

func (adapter *WeedAdapter) GetUrl(fid string) (url string, err error) {
	if len(fid) < 5 {
		return "", InvalidFid
	}
	index := strings.Index(fid, ",")
	if index == -1 && index == 0 {
		return "", InvalidFid
	}
	volumeId, err := strconv.Atoi(fid[:index])
	if err != nil {
		return "", InvalidFid
	}

	volumeUrl, ok := adapter.volumes[volumeId]

	if !ok {
		v, err := adapter.client.Volume(fid[:index], "")
		if err != nil {
			return "", err
		}
		volumeUrl = v.PublicUrl
		adapter.volumes[volumeId] = volumeUrl
	}
	url = volumeUrl + "/" + fid
	return
}
