package cases

// 一次只测一种策略组合，其它的注释掉。需至少启用一个 section。

// GenerateTestCases 生成策略用例。
//
// 每种策略组合单独对应一个 case_multi_*.go 文件，可单独启用/禁用以测试该组合。
// 双因子、三因子均做参数+权重网格探索，找各组合的最优参数。
func GenerateTestCases() []TestCase {
	b := &caseBuilder{cases: make([]TestCase, 0, 200)}

	// 单因子基线（约45组）
	// b.addBollSections()
	// b.addRsiSingleSections()
	// b.addBreakoutSections()

	// 双因子：精简网格，单次合计≤200组（可注释不需要的组合）
	// b.addMultiBollBreakSections()
	// b.addMultiBollRsiSections()
	// b.addMultiBollAtrSections()
	// b.addMultiBollMaSections()
	// b.addMultiBollTrendSections()
	// b.addMultiRsiBreakSections()
	// b.addMultiBreakAtrSections()
	// b.addMultiBollPriceVmaSections()  // 按需启用
	// b.addMultiBollVolumeSections()
	// b.addMultiBollMacdSections()
	// b.addMultiBollMacrossSections()
	// b.addMultiRsiAtrSections()
	// b.addMultiMaTrendSections()

	// 三因子：精简网格（可分批启用，每次≤200组）
	// b.addMultiTripleRsiBollBreakSections()
	// b.addMultiTripleBollBreakAtrSections()
	// b.addMultiTripleRsiBreakAtrSections()
	// b.addMultiTripleBollRsiAtrSections()
	// b.addMultiTripleBollBreakMaSections()
	// b.addMultiTripleBollBreakTrendSections()

	// 四因子：基于最优三因子 Bo+R+Br 叠加第4因子，每次只测一种（约100组）
	// b.addMultiQuadRsiBollBreakAtrSections()
	// b.addMultiQuadRsiBollBreakMaSections()
	// b.addMultiQuadRsiBollBreakTrendSections()
	// b.addMultiQuadRsiBollBreakSessionSections()
	// b.addMultiQuadRsiBollBreakPriceVmaSections()

	// 五因子：一次只测一种组合，约100组（R+Boll+Br+A 叠加第5因子）
	// b.addMultiQuintRsiBollBreakAtrSessSections()
	// b.addMultiQuintRsiBollBreakAtrTrendSections()
	// b.addMultiQuintRsiBollBreakAtrMaSections()
	// b.addMultiQuintRsiBollBreakAtrPvSections()
	// b.addMultiQuintRsiBollBreakAtrVolSections()  // 五因子最优 53.86%

	// 默认：单因子 Boll（若五因子全注释则启用）
	b.addBollSections()

	return b.result()
}
