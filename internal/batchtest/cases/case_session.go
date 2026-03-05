package cases

// addSessionSections Session 单因子（FactorSession 已输出正向信号）
func (b *caseBuilder) addSessionSections() {
	b.add("Session", b.session(1))
}
