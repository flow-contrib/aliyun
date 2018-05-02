package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

func (p *Aliyun) AddDomainRecord() (err error) {

	dnsListConf := p.Config.GetConfig("aliyun.dns")

	if dnsListConf.IsEmpty() {
		return
	}

	for _, dnsConfName := range dnsListConf.Keys() {

		dnsConf := dnsListConf.GetConfig(dnsConfName)

		if dnsConf.IsEmpty() {
			return
		}

		req := alidns.CreateAddDomainRecordRequest()

		req.DomainName = dnsConf.GetString("domain-name")
		req.RR = dnsConf.GetString("rr")
		req.Type = dnsConf.GetString("type")
		req.Value = dnsConf.GetString("value")
		req.TTL = requests.Integer(dnsConf.GetString("ttl", "600"))
		req.Priority = requests.Integer(dnsConf.GetString("priority", "10"))
		req.Line = dnsConf.GetString("line", "default")

		_, err = p.DNSClient().AddDomainRecord(req)

		if err != nil {
			return
		}
	}

	p.DNSClient().AddBatchDomainRecords(request)

	return
}

func (p *Aliyun) UpdateDomainRecord() (err error) {

	dnsListConf := p.Config.GetConfig("aliyun.dns")

	if dnsListConf.IsEmpty() {
		return
	}

	for _, dnsConfName := range dnsListConf.Keys() {

		dnsConf := dnsListConf.GetConfig(dnsConfName)

		if dnsConf.IsEmpty() {
			return
		}

		describeReq := alidns.CreateDescribeDomainRecordsRequest()

		describeReq.DomainName = dnsConf.GetString("domain-name")

		var describeResp *alidns.DescribeDomainRecordsResponse

		describeResp, err = p.DNSClient().DescribeDomainRecords(describeReq)

		if err != nil {
			return
		}

		req := alidns.CreateUpdateDomainRecordRequest()

		req.RR = dnsConf.GetString("rr")
		req.Type = dnsConf.GetString("type")
		req.Value = dnsConf.GetString("value")
		req.TTL = requests.Integer(dnsConf.GetString("ttl", "600"))
		req.Priority = requests.Integer(dnsConf.GetString("priority", "10"))
		req.Line = dnsConf.GetString("line", "default")

		if len(describeResp.DomainRecords.Record) == 0 {
			continue
		}

		record := describeResp.DomainRecords.Record[0]

		if record.RR == req.RR &&
			record.Type == req.Type &&
			record.Value == req.Value &&
			requests.Integer(record.TTL) == req.TTL &&
			requests.Integer(record.Priority) == req.Priority &&
			record.Line == req.Line {
			continue
		}

		req.RecordId = record.RecordId

		_, err = p.DNSClient().UpdateDomainRecord(req)

		if err != nil {
			return
		}
	}

	return
}

func (p *Aliyun) DeleteDomainRecord() (err error) {

	dnsListConf := p.Config.GetConfig("aliyun.dns")

	if dnsListConf.IsEmpty() {
		return
	}

	for _, dnsConfName := range dnsListConf.Keys() {

		dnsConf := dnsListConf.GetConfig(dnsConfName)

		if dnsConf.IsEmpty() {
			return
		}

		describeReq := alidns.CreateDescribeDomainRecordsRequest()

		describeReq.DomainName = dnsConf.GetString("domain-name")

		var describeResp *alidns.DescribeDomainRecordsResponse

		describeResp, err = p.DNSClient().DescribeDomainRecords(describeReq)

		if err != nil {
			return
		}

		if len(describeResp.DomainRecords.Record) == 0 {
			continue
		}

		record := describeResp.DomainRecords.Record[0]

		req := alidns.CreateDeleteDomainRecordRequest()
		req.RecordId = record.RecordId

		_, err = p.DNSClient().DeleteDomainRecord(req)

		if err != nil {
			return
		}
	}

	return
}
