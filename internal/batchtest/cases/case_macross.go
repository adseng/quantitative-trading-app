package cases

import "fmt"

// addMacrossSections 金叉/死叉（事件型+时间容错+预判）
func (b *caseBuilder) addMacrossSections() {
	// SECTION 12b: 金叉/死叉
	for _, params := range [][3]int{
		{5, 10, 1}, {5, 20, 1}, {5, 30, 1},
		{10, 20, 1}, {10, 30, 1}, {15, 30, 1}, {20, 30, 1},
		{20, 60, 1}, {30, 60, 1}, {50, 120, 1}, {50, 200, 1},
	} {
		b.add(fmt.Sprintf("MACr%d_%d_w2", params[0], params[1]), TestCase{
			UseMACross: true, MACrossShort: params[0], MACrossLong: params[1], MACrossWeight: float64(params[2]),
			MACrossWindow: 2, MACrossPreempt: 0,
		})
	}
	for _, params := range [][3]int{
		{5, 20, -1}, {10, 30, -1}, {20, 60, -1},
	} {
		b.add(fmt.Sprintf("MACr%d_%d_neg", params[0], params[1]), TestCase{
			UseMACross: true, MACrossShort: params[0], MACrossLong: params[1], MACrossWeight: float64(params[2]),
			MACrossWindow: 2, MACrossPreempt: 0,
		})
	}
	b.add("MACr5_20_w2p", TestCase{UseMACross: true, MACrossShort: 5, MACrossLong: 20, MACrossWeight: 1, MACrossWindow: 2, MACrossPreempt: 0.002})
	b.add("MACr10_30_w3p", TestCase{UseMACross: true, MACrossShort: 10, MACrossLong: 30, MACrossWeight: 1, MACrossWindow: 3, MACrossPreempt: 0.002})
	b.add("MACr5_20_neg_w2p", TestCase{UseMACross: true, MACrossShort: 5, MACrossLong: 20, MACrossWeight: -1, MACrossWindow: 2, MACrossPreempt: 0.002})
}
