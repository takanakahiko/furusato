package furusato

import (
	"fmt"
	"math"
)

const (
	// residentTaxRate is 住民税率.
	// 横浜市: 市民税8% + 県民税2.025%(2% + 水源環境保全税0.025%) = 10.025%
	residentTaxRate = 0.10025

	// specialIncomeTaxRateForReconstruction is 復興特別所得税率.
	specialIncomeTaxRateForReconstruction = 0.021

	// residentTaxBasicDeduction is 住民税の基礎控除.
	residentTaxBasicDeduction = 430_000
)

// TaxType は税区分を表す.
type TaxType int

const (
	// IncomeTaxType は所得税.
	IncomeTaxType TaxType = iota
	// ResidentTaxType は住民税.
	ResidentTaxType
)

// incomeTaxBasicDeduction は所得税の基礎控除を計算する.
// 令和7年度税制改正により、合計所得金額に応じて段階的に設定.
// https://www.nta.go.jp/users/gensen/2025kiso/index.htm
func incomeTaxBasicDeduction(totalIncome int) int {
	switch {
	case totalIncome <= 1_320_000:
		return 950_000
	case totalIncome <= 3_360_000:
		return 880_000
	case totalIncome <= 4_890_000:
		return 680_000
	case totalIncome <= 6_550_000:
		return 630_000
	case totalIncome <= 23_500_000:
		return 580_000
	case totalIncome <= 24_000_000:
		return 480_000
	case totalIncome <= 24_500_000:
		return 320_000
	case totalIncome <= 25_000_000:
		return 160_000
	default:
		return 0
	}
}

// BlueDeduction is 青色申告控除額.
func BlueDeduction(method DeclarationMethod, businessIncome int) int {
	var baseDeduction int

	switch method {
	case NoneDeclaration:
		baseDeduction = 0
	case ElectronicDeclaration:
		baseDeduction = 650_000
	case PaperDeclaration:
		baseDeduction = 550_000
	case SimpleDeclaration:
		baseDeduction = 100_000
	default:
		baseDeduction = 0

		fmt.Println("不正な申請方法が指定されました。青色申告控除なしを適用します。")
	}

	if businessIncome < baseDeduction {
		return businessIncome
	}

	return baseDeduction
}

// salaryIncomeDeduction は給与所得控除額を計算する.
// 令和7年度税制改正により、最低保障額が55万円→65万円に引き上げ.
// https://www.nta.go.jp/users/gensen/2025kiso/index.htm
func salaryIncomeDeduction(income int) int {
	switch {
	case income <= 1_900_000:
		return 650_000
	case income <= 3_600_000:
		return int(float64(income)*0.3) + 80_000
	case income <= 6_600_000:
		return int(float64(income)*0.2) + 440_000
	case income <= 8_500_000:
		return int(float64(income)*0.1) + 1_100_000
	default:
		return 1_950_000
	}
}

// TotalIncome is 総所得(給与所得控除済み).
func TotalIncome(input TaxCalculationInput) int {
	return input.SalaryIncome - salaryIncomeDeduction(input.SalaryIncome) +
		input.MiscellaneousIncome + input.BusinessIncome
}

// EarthquakeInsuranceDeduction は地震保険料控除額を計算する.
// 所得税: 支払額全額（上限50,000円）
// 住民税: 支払額の1/2（上限25,000円）
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/1145.htm
func EarthquakeInsuranceDeduction(input TaxCalculationInput, taxType TaxType) int {
	var deduction, maxDeduction int

	switch taxType {
	case IncomeTaxType:
		deduction = input.EarthquakeInsurance
		maxDeduction = 50_000
	case ResidentTaxType:
		deduction = input.EarthquakeInsurance / 2
		maxDeduction = 25_000
	default:
		return 0
	}

	return min(deduction, maxDeduction)
}

// lifeInsuranceDeductionPerCategory は生命保険料控除の各区分の控除額を計算する.
// 新制度（2012年1月1日以降の契約）の計算式を使用.
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/1140.htm
func lifeInsuranceDeductionPerCategory(premium int, taxType TaxType) int {
	switch taxType {
	case IncomeTaxType:
		// 所得税の計算式（新制度）
		// 〜20,000円: 全額
		// 20,001〜40,000円: 支払額×1/2 + 10,000円
		// 40,001〜80,000円: 支払額×1/4 + 20,000円
		// 80,001円〜: 一律40,000円
		switch {
		case premium <= 20_000:
			return premium
		case premium <= 40_000:
			return premium/2 + 10_000
		case premium <= 80_000:
			return premium/4 + 20_000
		default:
			return 40_000
		}
	case ResidentTaxType:
		// 住民税の計算式（新制度）
		// 〜12,000円: 全額
		// 12,001〜32,000円: 支払額×1/2 + 6,000円
		// 32,001〜56,000円: 支払額×1/4 + 14,000円
		// 56,001円〜: 一律28,000円
		switch {
		case premium <= 12_000:
			return premium
		case premium <= 32_000:
			return premium/2 + 6_000
		case premium <= 56_000:
			return premium/4 + 14_000
		default:
			return 28_000
		}
	default:
		return 0
	}
}

// LifeInsuranceDeduction は生命保険料控除の合計額を計算する.
// 一般生命保険料、介護医療保険料、個人年金保険料の3区分の合計.
// 所得税: 各区分上限40,000円、合計上限120,000円
// 住民税: 各区分上限28,000円、合計上限70,000円
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/1140.htm
func LifeInsuranceDeduction(input TaxCalculationInput, taxType TaxType) int {
	generalDeduction := lifeInsuranceDeductionPerCategory(input.LifeInsuranceGeneral, taxType)
	medicalDeduction := lifeInsuranceDeductionPerCategory(input.LifeInsuranceMedical, taxType)
	pensionDeduction := lifeInsuranceDeductionPerCategory(input.LifeInsurancePension, taxType)

	total := generalDeduction + medicalDeduction + pensionDeduction

	// 合計上限額
	var maxTotal int
	switch taxType {
	case IncomeTaxType:
		maxTotal = 120_000
	case ResidentTaxType:
		maxTotal = 70_000
	default:
		return 0
	}

	return min(total, maxTotal)
}

// MedicalDeduction is 医療費控除.
func MedicalDeduction(input TaxCalculationInput) int {
	// 医療費控除
	totalIncome := TotalIncome(input)

	// 医療費控除のしきい値（所得の5%または10万円のいずれか低い方）
	threshold := int(math.Min(float64(totalIncome)*0.05, 100_000))

	// 医療費控除
	medicalDeduction := input.MedicalExpenses - threshold
	if medicalDeduction < 0 {
		return 0
	}

	return medicalDeduction
}

// TaxableIncomeForIncomeTax is 所得税にかかる課税所得.
func TaxableIncomeForIncomeTax(input TaxCalculationInput) int {
	return TaxableIncome(input, IncomeTaxType)
}

// TaxableIncomeForResindentTax is 住民税にかかる課税所得.
func TaxableIncomeForResindentTax(input TaxCalculationInput) int {
	return TaxableIncome(input, ResidentTaxType)
}

// TaxableIncome is 課税所得.
func TaxableIncome(input TaxCalculationInput, taxType TaxType) int {
	// 基礎控除
	var basicDeduction int
	if taxType == IncomeTaxType {
		basicDeduction = incomeTaxBasicDeduction(TotalIncome(input))
	} else {
		basicDeduction = residentTaxBasicDeduction
	}

	// 医療費控除
	medicalDeduction := MedicalDeduction(input)

	// 申告特別控除
	blueDeduction := BlueDeduction(input.DeclarationMethod, input.BusinessIncome)

	// 扶養控除
	dependentDeduction := input.DependentCount * 380_000 // TODO 特定扶養親族などの分は別途計算が必要

	// 配偶者控除
	spouseDeduction := 0
	if input.SpouseDeduction {
		spouseDeduction = 380_000
	}

	// 地震保険料控除
	earthquakeDeduction := EarthquakeInsuranceDeduction(input, taxType)

	// 生命保険料控除
	lifeInsuranceDeduction := LifeInsuranceDeduction(input, taxType)

	// 課税所得
	taxableIncome := TotalIncome(input) - medicalDeduction -
		blueDeduction - input.SocialInsurance - dependentDeduction - spouseDeduction -
		earthquakeDeduction - lifeInsuranceDeduction - basicDeduction
	if taxableIncome < 0 {
		return 0
	}

	return taxableIncome
}

// 所得税率と控除額(所得税には復興特別所得税を含まない)
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/2260.htm
func CalculateIncomeTaxRate(input TaxCalculationInput) (float64, int) {
	taxableIncome := TaxableIncome(input, IncomeTaxType)

	switch {
	case taxableIncome <= 1_949_000:
		return 0.05, 0
	case taxableIncome <= 3_299_999:
		return 0.10, 97_500
	case taxableIncome <= 6_949_999:
		return 0.20, 427_500
	case taxableIncome <= 8_999_999:
		return 0.23, 636_000
	case taxableIncome <= 17_999_999:
		return 0.33, 1_536_000
	case taxableIncome <= 39_999_999:
		return 0.40, 2_796_000
	default:
		return 0.45, 4_796_000
	}
}

// IncomeTax is 所得税(復興特別所得税を含む).
func IncomeTax(input TaxCalculationInput) int {
	// 課税所得
	taxableIncome := TaxableIncome(input, IncomeTaxType)

	// ふるさと納税の控除
	// 所得税からの控除をする場合は税から控除するのではなく、課税所得から控除する（後の短数切り捨てに巻き込まれる）
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/04_3.htm
	if input.furusatoAmount > 0 {
		taxableIncome -= (input.furusatoAmount - 2000)
	}

	// 課税所得金額を1000円未満の端数切り捨て
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/02_1.htm
	// > 1,000円未満切捨て
	taxableIncome = taxableIncome / 1000 * 1000 // 1000円未満の端数切り捨て

	// 特別所得税を含めた所得税率と控除額
	rate, deduction := CalculateIncomeTaxRate(input)

	incomeTax := int(float64(taxableIncome)*rate) - deduction
	if incomeTax < 0 {
		return 0
	}

	incomeTax += int(float64(incomeTax) * specialIncomeTaxRateForReconstruction)

	// 定額減税
	// incomeTax = incomeTax - 30000

	// 100円未満の端数切り捨て
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/01_1.htm
	// > 100円未満切捨て
	incomeTax = incomeTax / 100 * 100 // 100円未満の端数切り捨て

	return incomeTax
}

// HousingLoanDeductionForResidentTax は住民税から控除できる住宅ローン控除額を計算する.
// 所得税から控除しきれなかった額を住民税から控除できる（上限あり）.
// https://www.soumu.go.jp/main_sosiki/jichi_zeisei/czaisei/czaisei_seido/090929.html
func HousingLoanDeductionForResidentTax(input TaxCalculationInput) int {
	if input.HousingLoanDeduction == 0 {
		return 0
	}

	// 所得税額（住宅ローン控除前）
	incomeTaxBeforeHousing := IncomeTaxBeforeHousingLoan(input)

	// 所得税から控除しきれない額
	remainingDeduction := input.HousingLoanDeduction - incomeTaxBeforeHousing
	if remainingDeduction <= 0 {
		return 0
	}

	// 住民税からの控除上限額
	// 所得税の課税所得金額×7%（最高136,500円）
	// ※2014年4月以降入居で消費税8%または10%の場合
	taxableIncome := TaxableIncome(input, IncomeTaxType)
	maxDeduction := int(float64(taxableIncome) * 0.07)
	if maxDeduction > 136_500 {
		maxDeduction = 136_500
	}

	if remainingDeduction > maxDeduction {
		return maxDeduction
	}

	return remainingDeduction
}

// IncomeTaxBeforeHousingLoan は住宅ローン控除前の所得税額を計算する.
func IncomeTaxBeforeHousingLoan(input TaxCalculationInput) int {
	// 課税所得
	taxableIncome := TaxableIncome(input, IncomeTaxType)

	// ふるさと納税の控除
	if input.furusatoAmount > 0 {
		taxableIncome -= (input.furusatoAmount - 2000)
	}

	// 課税所得金額を1000円未満の端数切り捨て
	taxableIncome = taxableIncome / 1000 * 1000

	// 所得税率と控除額
	rate, deduction := CalculateIncomeTaxRate(input)

	incomeTax := int(float64(taxableIncome)*rate) - deduction
	if incomeTax < 0 {
		return 0
	}

	incomeTax += int(float64(incomeTax) * specialIncomeTaxRateForReconstruction)

	// 100円未満の端数切り捨て
	incomeTax = incomeTax / 100 * 100

	return incomeTax
}

// ResidentTax is 住民税所得割額(均等割額は含まない).
func ResidentTax(input TaxCalculationInput, noFurusato bool) int {
	// TODO: 調整控除は面倒で計算していません
	// https://www.city.itabashi.tokyo.jp/tetsuduki/zei/kuminzei/1001751.html
	adjastmentDeduction := 2500 // 調整控除

	// 課税所得
	taxableIncome := TaxableIncome(input, ResidentTaxType)

	// 住民税の課税標準額
	// 1000円未満の端数切り捨て
	// https://www.city.tokyo-nakano.lg.jp/kurashi/zeikin/jyuminzei-kazei/jyuminzei-keisanrei.html
	// > 課税標準額は、所得合計額から所得控除合計額を差し引いた額（1,000円未満切り捨て）です。
	taxableIncome = taxableIncome / 1000 * 1000

	// 住民税所得割額
	residentTax := int(float64(taxableIncome)*residentTaxRate) - adjastmentDeduction

	// 住宅ローン控除（住民税からの控除分）
	housingDeduction := HousingLoanDeductionForResidentTax(input)
	residentTax -= housingDeduction

	// ふるさと納税の控除
	// 住民税からの控除をする場合は所得から控除するのではなく、税から控除する
	// https://www.city.yokohama.lg.jp/kurashi/koseki-zei-hoken/zeikin/y-shizei/kojin-shiminzei-kenminzei/kojin-shiminzei-shosai/zeigakukoujo.html
	//
	//nolint:lll
	if input.furusatoAmount > 0 && !noFurusato {
		incomeTaxRate, _ := CalculateIncomeTaxRate(input)
		incomeTaxRateWithForReconstruction := incomeTaxRate * (1 + specialIncomeTaxRateForReconstruction)
		residentTax -= int(float64(input.furusatoAmount-2000) * residentTaxRate)

		// TODO: 限度額を超えた場合のロジックを書いたほうがいい
		residentTax -= int(float64(input.furusatoAmount-2000) * (1.0 - residentTaxRate - incomeTaxRateWithForReconstruction))
	}

	residentTax = residentTax / 100 * 100 // 100円未満の端数切り捨て

	return residentTax
}

func (input TaxCalculationInput) FurusatoDeductionOfIncomeTax(furusatoAmount int) int {
	input2 := input
	input2.furusatoAmount = furusatoAmount

	return IncomeTax(input) - IncomeTax(input2)
}

func (input TaxCalculationInput) FurusatoDeductionOfResidentTax(furusatoAmount int) int {
	input2 := input
	input2.furusatoAmount = furusatoAmount

	return ResidentTax(input, true) - ResidentTax(input2, false)
}

// FurusatoNozeiLimit is ふるさと納税の控除上限額.
func FurusatoNozeiLimit(input TaxCalculationInput) int {
	// 所得税率と控除額
	incomeTaxRate, _ := CalculateIncomeTaxRate(input)
	incomeTaxRateWithForReconstruction := incomeTaxRate * (1 + specialIncomeTaxRateForReconstruction)

	// 住民税
	residentTax := ResidentTax(input, true) // 住民税所得割額

	// ふるさと納税の控除上限額
	// https://www.soumu.go.jp/main_sosiki/jichi_zeisei/czaisei/czaisei_seido/furusato/mechanism/deduction.html
	// 住民税特例分からの控除限度額＝個人住民税所得割額の20%
	// 住民税からの控除（特例分） = （ふるさと納税額 - 2,000円）×（100％ - 10％（基本分） - 所得税の税率）
	// 個人住民税所得割額 * 20％ = (X-2,000円) * (90％-所得税の税率)
	// X = 個人住民税所得割額 * 20％ /（90％-所得税の税率）+ 2,000円
	return int(float64(residentTax)*0.2/(1-residentTaxRate-incomeTaxRateWithForReconstruction)) + 2000
}
