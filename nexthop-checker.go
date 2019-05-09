package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
	"gopkg.in/cheggaaa/pb.v1"
)

// VrouterIntrospectCli struct
type VrouterIntrospectCli struct {
	host     *string
	port     *int
	NHList   []KNHInfo
	progress *bool
}

// Init for VrouterIntrospectCli
func (vip *VrouterIntrospectCli) Init(host string, port int, progress bool) {
	vip.host = &host
	vip.port = &port
	vip.progress = &progress

}

// Flow XML Structures

// FlowStatsRecordsResp struct
type FlowStatsRecordsResp struct {
	RecordsList RecordsList `xml:"records_list"`
}

// RecordsList struct
type RecordsList struct {
	List RList `xml:"list"`
}

// RList struct
type RList struct {
	FlowStatsRecord []FlowStatsRecord `xml:"FlowStatsRecord"`
}

// FlowStatsRecord struct
type FlowStatsRecord struct {
	FSRInfo FSRInfo `xml:"info"`
}

// FSRInfo struct
type FSRInfo struct {
	SandeshFlowExportInfo SandeshFlowExportInfo `xml:"SandeshFlowExportInfo"`
}

// SandeshFlowExportInfo struct
type SandeshFlowExportInfo struct {
	Key        Key    `xml:"key"`
	EgressUUID string `xml:"egress_uuid"`
}

// Key struct
type Key struct {
	SandeshFlowKey SandeshFlowKey `xml:"SandeshFlowKey"`
}

// SandeshFlowKey struct
type SandeshFlowKey struct {
	NH       uint32 `xml:"nh"`
	Sip      string `xml:"sip"`
	Dip      string `xml:"dip"`
	SrcPort  uint32 `xml:"src_port"`
	DstPort  uint32 `xml:"dst_port"`
	Protocol uint16 `xml:"protocol"`
}

// CheckFlows for VrouterIntrospectCli
func (vip *VrouterIntrospectCli) CheckFlows() {

	url := fmt.Sprintf("http://%s:%d/Snh_FlowStatsRecordsReq", *vip.host, *vip.port)
	resp, err := http.Get(url)
	if err != nil {
		glog.Error(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
	}

	vip.getNextHopList() // load list of nexthops
	var fsrr FlowStatsRecordsResp
	xml.Unmarshal(body, &fsrr)
	var bar *pb.ProgressBar
	if *vip.progress == true {
		bar = pb.StartNew(len(fsrr.RecordsList.List.FlowStatsRecord))
	}
	for _, flow := range fsrr.RecordsList.List.FlowStatsRecord {
		knh := vip.getNextHopByID(flow.FSRInfo.SandeshFlowExportInfo.Key.SandeshFlowKey.NH)

		if knh != nil {
			if glog.V(2) {
				sfk := flow.FSRInfo.SandeshFlowExportInfo.Key.SandeshFlowKey
				jsn, _ := json.Marshal(sfk)
				glog.Infof("NextHop exists for flow: %s ", jsn)
				if glog.V(3) {
					jsnh, _ := json.Marshal(knh)
					glog.Infof("NextHop: %s ", jsnh)
				}
			}
		} else {
			sfk := flow.FSRInfo.SandeshFlowExportInfo.Key.SandeshFlowKey
			jsn, _ := json.Marshal(sfk)
			if glog.V(1) {
				glog.Warningf("NextHop doesn't exist for flow: %s", jsn)
			}
		}
		if bar != nil {
			bar.Increment()
		}
	}
	if bar != nil {
		bar.Finish()
	}
}

// NextHop XML Structures

// KNHRespList struct
type KNHRespList struct {
	List []KNHResp `xml:"KNHResp"`
}

// KNHResp struct
type KNHResp struct {
	NHList NHList `xml:"nh_list"`
}

// NHList struct
type NHList struct {
	List NHInterList `xml:"list"`
}

// NHInterList struct
type NHInterList struct {
	KNHInfo []KNHInfo `xml:"KNHInfo"`
}

// KNHInfo struct
type KNHInfo struct {
	ID          uint32 `xml:"id"`
	Type        string `xml:"type"`
	VRF         int    `xml:"vrf"`
	Flag        string `xml:"flags"`
	EncapFamily string `xml:"encap_family"`
	EncapOif    string `xml:"encap_oif_id"`
}

// ErrResp struct
type ErrResp struct {
	Resp string `xml:"resp"`
}

func (vip *VrouterIntrospectCli) getNextHopByID(id uint32) *KNHInfo {

	for _, knh := range vip.NHList {
		if knh.ID == id {
			return &knh
		}
	}
	return nil
}

func (vip *VrouterIntrospectCli) getNextHopList() {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/Snh_KNHReq?nh_id=", *vip.host, *vip.port))
	if err != nil {
		glog.Error(err)

	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
	}
	var knhList KNHRespList
	err = xml.Unmarshal(body, &knhList)
	if err != nil {
		glog.Error(err)
	}
	for _, knhResp := range knhList.List {
		for _, knh := range knhResp.NHList.List.KNHInfo {
			vip.NHList = append(vip.NHList, knh)
			if glog.V(3) {
				jsn, _ := json.Marshal(knh)
				glog.Infof("Loading NH: %s", string(jsn))
			}
		}
	}
}

// Usage func
func Usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	fmt.Printf("It loads all NextHops and Flows from vRouter Introspect and checks if NextHop with ID from Flow really exists .\n")
	fmt.Println("\n Flags:")
	flag.PrintDefaults()
	fmt.Printf("\n Verbose levels:")
	fmt.Printf("\n\t Error\t\t--v 0")
	fmt.Printf("\n\t Warning \t--v 1 (default)")
	fmt.Printf("\n\t Info \t\t--v 2")
	fmt.Printf("\n\t Debug mode \t--v 3 \n")
}

func main() {

	hostPtr := flag.String("host", "127.0.0.1", "vRouter IP address")
	portPtr := flag.Int("port", 8085, "vRouter introspection port")
	progressPtr := flag.Bool("progress", false, "Show progress bar")
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "WARNING")
	flag.Set("v", "1")
	flag.Usage = func() { Usage() }
	flag.Parse()

	vip := VrouterIntrospectCli{host: hostPtr, port: portPtr, progress: progressPtr}
	glog.Infof("Start checking ...")
	vip.CheckFlows()
	glog.Flush()
	glog.Infof("Done")
}
