package furusato

import (
	"fmt"
	"math"
)

const (
	// residentTaxRate is 住民税率.
	residentTaxRate = 0.10025 // 神奈川県

	// specialIncomeTaxRateForReconstruction is 復興特別所得税率.
	specialIncomeTaxRateForReconstruction = 0.021

	// incomeTaxBasicDeduction is 所得税の基礎控除.
	incomeTaxBasicDeduction = 480_000

	// residentTaxBasicDeduction is 住民税の基礎控除.
	residentTaxBasicDeduction = 430_000
)

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

// salaryIncomeDeduction is 給与所得控除額.
func salaryIncomeDeduction(income int) int {
	switch {
	case income <= 1_625_000:
		return 550_000
	case income <= 1_800_000:
		return int(float64(income)*0.4) - 100_000
	case income <= 3_600_000:
		return int(float64(income)*0.3) - 80_000
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
	return TaxableIncome(input, incomeTaxBasicDeduction)
}

// TaxableIncomeForResindentTax is 住民税にかかる課税所得.
func TaxableIncomeForResindentTax(input TaxCalculationInput) int {
	return TaxableIncome(input, residentTaxBasicDeduction)
}

// TaxableIncome is 課税所得.
func TaxableIncome(input TaxCalculationInput, basicDeduction int) int {
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

	// 課税所得
	taxableIncome := TotalIncome(input) - medicalDeduction -
		blueDeduction - input.SocialInsurance - dependentDeduction - spouseDeduction - basicDeduction
	if taxableIncome < 0 {
		return 0
	}

	return taxableIncome
}

// 所得税率と控除額(所得税には復興特別所得税を含まない)
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/2260.htm
func CalculateIncomeTaxRate(input TaxCalculationInput, basicDeduction int) (float64, int) {
	taxableIncome := TaxableIncome(input, basicDeduction)

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
	taxableIncome := TaxableIncome(input, incomeTaxBasicDeduction)

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
	rate, deduction := CalculateIncomeTaxRate(input, incomeTaxBasicDeduction)

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

// ResidentTax is 住民税所得割額(均等割額は含まない).
func ResidentTax(input TaxCalculationInput, noFurusato bool) int {
	// TODO: 調整控除は面倒で計算していません
	// https://www.city.itabashi.tokyo.jp/tetsuduki/zei/kuminzei/1001751.html
	adjastmentDeduction := 2500 // 調整控除

	// 課税所得
	taxableIncome := TaxableIncome(input, residentTaxBasicDeduction)

	// 住民税の課税標準額
	// 1000円未満の端数切り捨て
	// https://www.city.tokyo-nakano.lg.jp/kurashi/zeikin/jyuminzei-kazei/jyuminzei-keisanrei.html
	// > 課税標準額は、所得合計額から所得控除合計額を差し引いた額（1,000円未満切り捨て）です。
	taxableIncome = taxableIncome / 1000 * 1000

	// 住民税所得割額
	residentTax := int(float64(taxableIncome)*residentTaxRate) - adjastmentDeduction

	// ふるさと納税の控除
	// 住民税からの控除をする場合は所得から控除するのではなく、税から控除する
	// https://www.city.yokohama.lg.jp/kurashi/koseki-zei-hoken/zeikin/y-shizei/kojin-shiminzei-kenminzei/kojin-shiminzei-shosai/zeigakukoujo.html
	//
	//nolint:lll
	if input.furusatoAmount > 0 && !noFurusato {
		incomeTaxRate, _ := CalculateIncomeTaxRate(input, incomeTaxBasicDeduction)
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
	incomeTaxRate, _ := CalculateIncomeTaxRate(input, incomeTaxBasicDeduction)
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
