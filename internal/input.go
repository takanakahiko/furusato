package furusato

// declarationMethod is 申請方法.
type declarationMethod int

const (
	//  ElectronicDeclaration is 電子申告+電子帳簿保存.
	ElectronicDeclaration declarationMethod = iota

	// ElectronicDeclaration is 紙帳簿.
	PaperDeclaration

	// SimpleDeclaration is 簡易帳簿.
	SimpleDeclaration
)

// TaxCalculationInput is 入力データをまとめる構造体.
type TaxCalculationInput struct {
	// SalaryIncome is 給与収入(源泉徴収票の"支払金額")
	SalaryIncome int
	// MiscellaneousIncome is 雑所得
	MiscellaneousIncome int
	// BusinessIncome is 事業所得
	BusinessIncome int
	// MedicalExpenses is 医療費
	MedicalExpenses int
	// SocialInsurance is 社会保険料控除
	SocialInsurance int
	// DependentCount is 扶養親族数(一般の控除対象扶養親族)
	DependentCount int
	// SpouseDeduction is 配偶者控除の適用有無
	SpouseDeduction bool
	// Method is 申告方法
	Method declarationMethod

	// furusatoAmount is 内部的に減税効果を算出するためのふるさと納税額
	furusatoAmount int
}
