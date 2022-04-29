package mapper

import "snorba.art/hugo/dns-yml/util"

const (
	rtype_mx    = "MX"
	rtype_cname = "CNAME"
	rtype_a     = "A"
	rtype_aaaa  = "AAAA"
	rtype_dname = "DNAME"
	rtype_txt   = "TXT"
	rtype_srv   = "SRV"
	rtype_ptr   = "PTR"
	rtype_ns    = "NS"
	rtype_alias = "ALIAS"
	rtype_naptr = "NAPTR"
	rtype_tlsa  = "TLSA"
)

var rtypes []string = []string{
	rtype_mx,
	rtype_cname,
	rtype_a,
	rtype_aaaa,
	rtype_dname,
	rtype_txt,
	rtype_srv,
	rtype_ptr,
	rtype_ns,
	rtype_alias,
	rtype_naptr,
	rtype_tlsa,
}

func rTypeValid(rType string) bool {
	return util.SliceContainsString(rtypes, rType)
}
