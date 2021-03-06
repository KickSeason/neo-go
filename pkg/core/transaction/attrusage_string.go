// Code generated by "stringer -type=AttrUsage"; DO NOT EDIT.

package transaction

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ContractHash-0]
	_ = x[ECDH02-2]
	_ = x[ECDH03-3]
	_ = x[Vote-48]
	_ = x[CertURL-128]
	_ = x[DescriptionURL-129]
	_ = x[Description-144]
	_ = x[Hash1-161]
	_ = x[Hash2-162]
	_ = x[Hash3-163]
	_ = x[Hash4-164]
	_ = x[Hash5-165]
	_ = x[Hash6-166]
	_ = x[Hash7-167]
	_ = x[Hash8-168]
	_ = x[Hash9-169]
	_ = x[Hash10-170]
	_ = x[Hash11-171]
	_ = x[Hash12-172]
	_ = x[Hash13-173]
	_ = x[Hash14-174]
	_ = x[Hash15-175]
	_ = x[Remark-240]
	_ = x[Remark1-241]
	_ = x[Remark2-242]
	_ = x[Remark3-243]
	_ = x[Remark4-244]
	_ = x[Remark5-245]
	_ = x[Remark6-246]
	_ = x[Remark7-247]
	_ = x[Remark8-248]
	_ = x[Remark9-249]
	_ = x[Remark10-250]
	_ = x[Remark11-251]
	_ = x[Remark12-252]
	_ = x[Remark13-253]
	_ = x[Remark14-254]
	_ = x[Remark15-255]
}

const (
	_AttrUsage_name_0 = "ContractHash"
	_AttrUsage_name_1 = "ECDH02ECDH03"
	_AttrUsage_name_2 = "Vote"
	_AttrUsage_name_3 = "CertURLDescriptionURL"
	_AttrUsage_name_4 = "Description"
	_AttrUsage_name_5 = "Hash1Hash2Hash3Hash4Hash5Hash6Hash7Hash8Hash9Hash10Hash11Hash12Hash13Hash14Hash15"
	_AttrUsage_name_6 = "RemarkRemark1Remark2Remark3Remark4Remark5Remark6Remark7Remark8Remark9Remark10Remark11Remark12Remark13Remark14Remark15"
)

var (
	_AttrUsage_index_1 = [...]uint8{0, 6, 12}
	_AttrUsage_index_3 = [...]uint8{0, 7, 21}
	_AttrUsage_index_5 = [...]uint8{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 51, 57, 63, 69, 75, 81}
	_AttrUsage_index_6 = [...]uint8{0, 6, 13, 20, 27, 34, 41, 48, 55, 62, 69, 77, 85, 93, 101, 109, 117}
)

func (i AttrUsage) String() string {
	switch {
	case i == 0:
		return _AttrUsage_name_0
	case 2 <= i && i <= 3:
		i -= 2
		return _AttrUsage_name_1[_AttrUsage_index_1[i]:_AttrUsage_index_1[i+1]]
	case i == 48:
		return _AttrUsage_name_2
	case 128 <= i && i <= 129:
		i -= 128
		return _AttrUsage_name_3[_AttrUsage_index_3[i]:_AttrUsage_index_3[i+1]]
	case i == 144:
		return _AttrUsage_name_4
	case 161 <= i && i <= 175:
		i -= 161
		return _AttrUsage_name_5[_AttrUsage_index_5[i]:_AttrUsage_index_5[i+1]]
	case 240 <= i && i <= 255:
		i -= 240
		return _AttrUsage_name_6[_AttrUsage_index_6[i]:_AttrUsage_index_6[i+1]]
	default:
		return "AttrUsage(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
